package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/yourusername/clai/internal/llm"
)

// JSONStore is a file-based storage backend. Each session is a separate JSON file.
type JSONStore struct {
	baseDir string
}

// NewJSONStore creates a new JSONStore backed by the given directory.
func NewJSONStore(baseDir string) (*JSONStore, error) {
	sessionsDir := filepath.Join(baseDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &JSONStore{baseDir: baseDir}, nil
}

func (s *JSONStore) sessionPath(id string) string {
	return filepath.Join(s.baseDir, "sessions", id+".json")
}

func (s *JSONStore) writeSession(sess *Session) error {
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.sessionPath(sess.ID), data, 0644)
}

func (s *JSONStore) readSession(id string) (*Session, error) {
	data, err := os.ReadFile(s.sessionPath(id))
	if err != nil {
		return nil, err
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

// truncateToFirstMessage returns the first ~50 chars of the first user message
// to use as an auto-generated session name.
func truncateToFirstMessage(msg string) string {
	const max = 50
	if utf8.RuneCountInString(msg) <= max {
		return msg
	}
	runes := []rune(msg)
	return string(runes[:max]) + "…"
}

func (s *JSONStore) CreateSession(name string) (*Session, error) {
	now := time.Now()
	sess := &Session{
		ID:        uuid.New().String(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  []llm.Message{},
	}
	if err := s.writeSession(sess); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *JSONStore) GetSession(id string) (*Session, error) {
	return s.readSession(id)
}

func (s *JSONStore) ListSessions() ([]Session, error) {
	pattern := filepath.Join(s.baseDir, "sessions", "*.json")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var sess Session
		if err := json.Unmarshal(data, &sess); err != nil {
			continue
		}
		sessions = append(sessions, sess)
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

func (s *JSONStore) UpdateSession(session *Session) error {
	session.UpdatedAt = time.Now()
	return s.writeSession(session)
}

func (s *JSONStore) DeleteSession(id string) error {
	return os.Remove(s.sessionPath(id))
}

func (s *JSONStore) AddMessage(sessionID string, msg llm.Message) error {
	sess, err := s.readSession(sessionID)
	if err != nil {
		return fmt.Errorf("get session %s: %w", sessionID, err)
	}

	// Auto-name the session from the first user message
	if len(sess.Messages) == 0 && msg.Role == llm.RoleUser && sess.Name == "New Session" {
		sess.Name = truncateToFirstMessage(msg.Content)
	}

	sess.Messages = append(sess.Messages, msg)
	sess.UpdatedAt = time.Now()
	return s.writeSession(sess)
}

func (s *JSONStore) GetMessages(sessionID string) ([]llm.Message, error) {
	sess, err := s.readSession(sessionID)
	if err != nil {
		return nil, err
	}
	return sess.Messages, nil
}
