package storage

import (
	"time"

	"github.com/lioarce01/clai/internal/llm"
)

// Session is a named conversation thread.
type Session struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Messages  []llm.Message `json:"messages"`
}

// Store defines persistence operations for sessions and messages.
type Store interface {
	CreateSession(name string) (*Session, error)
	GetSession(id string) (*Session, error)
	ListSessions() ([]Session, error)
	UpdateSession(session *Session) error
	DeleteSession(id string) error

	AddMessage(sessionID string, msg llm.Message) error
	GetMessages(sessionID string) ([]llm.Message, error)
}
