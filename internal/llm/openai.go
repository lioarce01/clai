package llm

import (
	"context"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

type openAIClient struct {
	client *openai.Client
}

func newOpenAIClient(apiKey, baseURL string) *openAIClient {
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return &openAIClient{client: openai.NewClientWithConfig(cfg)}
}

// toOpenAIMessages converts our internal Message slice to the openai SDK format,
// prepending the system prompt if provided.
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

func (c *openAIClient) ChatCompletionStream(ctx context.Context, params CompletionParams) (<-chan StreamDelta, error) {
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       params.Model,
		Messages:    toOpenAIMessages(params),
		Temperature: float32(params.Temperature),
		MaxTokens:   params.MaxTokens,
		TopP:        float32(params.TopP),
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("create stream: %w", err)
	}

	ch := make(chan StreamDelta, 64)

	go func() {
		defer close(ch)
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					ch <- StreamDelta{Done: true}
				} else {
					ch <- StreamDelta{Error: err, Done: true}
				}
				return
			}

			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta.Content
				if delta != "" {
					select {
					case ch <- StreamDelta{Content: delta}:
					case <-ctx.Done():
						ch <- StreamDelta{Error: ctx.Err(), Done: true}
						return
					}
				}
			}
		}
	}()

	return ch, nil
}
