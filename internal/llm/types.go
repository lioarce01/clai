package llm

import "time"

// Role represents the role of a message in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a single turn in a conversation.
type Message struct {
	ID        string    `json:"id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Reasoning string    `json:"reasoning,omitempty"` // populated by reasoning models
	CreatedAt time.Time `json:"created_at"`
	Tokens    int       `json:"tokens,omitempty"`
}

// CompletionParams holds the parameters for a completion request.
type CompletionParams struct {
	Model        string
	Messages     []Message
	Temperature  float64
	MaxTokens    int
	TopP         float64
	SystemPrompt string
}

// StreamDelta is a single chunk delivered during streaming.
type StreamDelta struct {
	Content   string // Incremental content chunk
	Reasoning string // Incremental reasoning chunk (reasoning models only)
	Done      bool   // True when the stream is finished
	Error     error  // Non-nil if an error occurred
	Usage     *Usage // Token usage (only on the final chunk)
}

// Usage contains token consumption information.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
