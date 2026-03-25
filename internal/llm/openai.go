package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type openAIClient struct {
	client  *openai.Client
	apiKey  string
	baseURL string
	http    *http.Client
}

func newOpenAIClient(apiKey, baseURL string) *openAIClient {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return &openAIClient{
		client:  openai.NewClientWithConfig(cfg),
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{},
	}
}

func toOpenAIMessages(params CompletionParams) []openai.ChatCompletionMessage {
	msgs := make([]openai.ChatCompletionMessage, 0, len(params.Messages)+1)
	if params.SystemPrompt != "" {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    string(RoleSystem),
			Content: params.SystemPrompt,
		})
	}
	for _, m := range params.Messages {
		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	return msgs
}

func (c *openAIClient) ChatCompletion(ctx context.Context, params CompletionParams) (Message, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       params.Model,
		Messages:    toOpenAIMessages(params),
		Temperature: float32(params.Temperature),
		MaxTokens:   params.MaxTokens,
		TopP:        float32(params.TopP),
	})
	if err != nil {
		return Message{}, fmt.Errorf("chat completion: %w", err)
	}
	if len(resp.Choices) == 0 {
		return Message{}, fmt.Errorf("no completion choices returned")
	}
	return Message{
		Role:    RoleAssistant,
		Content: resp.Choices[0].Message.Content,
		Tokens:  resp.Usage.CompletionTokens,
	}, nil
}

// sseChunk is the raw structure of each SSE event from OpenAI-compatible APIs.
// Handles both standard (content) and reasoning model (reasoning) formats.
type sseChunk struct {
	Choices []struct {
		Delta struct {
			Content   *string `json:"content"`
			Reasoning string  `json:"reasoning"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *openAIClient) ChatCompletionStream(ctx context.Context, params CompletionParams) (<-chan StreamDelta, error) {
	type reqMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type reqBody struct {
		Model       string       `json:"model"`
		Messages    []reqMessage `json:"messages"`
		Temperature float64      `json:"temperature"`
		MaxTokens   int          `json:"max_tokens,omitempty"` // omitted when 0 — lets provider decide
		TopP        float64      `json:"top_p"`
		Stream      bool         `json:"stream"`
	}

	oaiMsgs := toOpenAIMessages(params)
	msgs := make([]reqMessage, len(oaiMsgs))
	for i, m := range oaiMsgs {
		msgs[i] = reqMessage{Role: m.Role, Content: m.Content}
	}

	bodyBytes, err := json.Marshal(reqBody{
		Model:       params.Model,
		Messages:    msgs,
		Temperature: params.Temperature,
		MaxTokens:   params.MaxTokens,
		TopP:        params.TopP,
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL+"/chat/completions",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(b))
	}

	ch := make(chan StreamDelta, 64)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		// Increase buffer for large chunks
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamDelta{Done: true}
				return
			}

			var chunk sseChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			var delta StreamDelta

			if chunk.Usage != nil {
				delta.Usage = &Usage{
					PromptTokens:     chunk.Usage.PromptTokens,
					CompletionTokens: chunk.Usage.CompletionTokens,
					TotalTokens:      chunk.Usage.TotalTokens,
				}
			}

			if len(chunk.Choices) > 0 {
				d := chunk.Choices[0].Delta
				if d.Content != nil {
					delta.Content = *d.Content
				}
				delta.Reasoning = d.Reasoning

				// Mark done on finish_reason but still deliver the delta
				if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != "" {
					delta.Done = true
				}
			}

			// Always send if there's anything to deliver
			if delta.Content != "" || delta.Reasoning != "" || delta.Usage != nil || delta.Done {
				select {
				case ch <- delta:
				case <-ctx.Done():
					ch <- StreamDelta{Error: ctx.Err(), Done: true}
					return
				}
				if delta.Done {
					return
				}
			}
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			ch <- StreamDelta{Error: err, Done: true}
		} else {
			ch <- StreamDelta{Done: true}
		}
	}()

	return ch, nil
}
