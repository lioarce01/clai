package llm

import "context"

// Client defines the contract for LLM API communication.
type Client interface {
	// ChatCompletion sends a blocking completion request and returns the full response.
	ChatCompletion(ctx context.Context, params CompletionParams) (Message, error)

	// ChatCompletionStream initiates a streaming request, returning a channel
	// that delivers incremental StreamDelta values. The channel is closed when
	// the stream ends (either Done=true or Error!=nil).
	ChatCompletionStream(ctx context.Context, params CompletionParams) (<-chan StreamDelta, error)
}

// NewClient creates the appropriate LLM client based on the provided configuration.
// Currently only OpenAI-compatible clients are supported.
func NewClient(apiKey, baseURL string) Client {
	return newOpenAIClient(apiKey, baseURL)
}
