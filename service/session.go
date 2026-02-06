package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tomohiro-owada/gmn-gui/internal/api"
)

// SessionSummary is the lightweight listing item (no messages)
type SessionSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Model     string    `json:"model"`
	WorkDir   string    `json:"workDir,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// SessionData is the full session stored on disk
type SessionData struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Model     string        `json:"model"`
	WorkDir   string        `json:"workDir,omitempty"`
	Messages  []ChatMessage `json:"messages"`
	History   []api.Content `json:"history"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// SessionService manages session persistence
type SessionService struct {
	ctx  context.Context
	chat *ChatService
	dir  string
}

// NewSessionService creates a new session service
func NewSessionService(chat *ChatService) *SessionService {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".gemini", "gmn-gui", "sessions")
	os.MkdirAll(dir, 0o755)

	return &SessionService{
		chat: chat,
		dir:  dir,
	}
}

// SetContext sets the Wails runtime context
func (s *SessionService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// ListSessions returns all session summaries sorted by updatedAt desc
func (s *SessionService) ListSessions() []SessionSummary {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return []SessionSummary{}
	}

	var sessions []SessionSummary
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}
		var sd SessionData
		if err := json.Unmarshal(data, &sd); err != nil {
			continue
		}
		sessions = append(sessions, SessionSummary{
			ID:        sd.ID,
			Title:     sd.Title,
			Model:     sd.Model,
			WorkDir:   sd.WorkDir,
			CreatedAt: sd.CreatedAt,
			UpdatedAt: sd.UpdatedAt,
		})
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions
}

// SaveCurrentSession saves the current chat state as a session
func (s *SessionService) SaveCurrentSession(id string) error {
	s.chat.mu.Lock()
	msgs := make([]ChatMessage, len(s.chat.messages))
	copy(msgs, s.chat.messages)
	hist := make([]api.Content, len(s.chat.history))
	copy(hist, s.chat.history)
	model := s.chat.model
	workDir := s.chat.workDir
	s.chat.mu.Unlock()

	if len(msgs) == 0 {
		return nil
	}

	// Generate title from first user message
	title := "New Chat"
	for _, m := range msgs {
		if m.Role == "user" {
			title = m.Content
			if len(title) > 60 {
				title = title[:60] + "..."
			}
			break
		}
	}

	now := time.Now()

	// Check if existing session
	existing := s.loadFile(id)
	createdAt := now
	if existing != nil {
		createdAt = existing.CreatedAt
	}

	sd := SessionData{
		ID:        id,
		Title:     title,
		Model:     model,
		WorkDir:   workDir,
		Messages:  msgs,
		History:   hist,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}

	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	return os.WriteFile(filepath.Join(s.dir, id+".json"), data, 0o644)
}

// LoadSession restores a session into the chat service
func (s *SessionService) LoadSession(id string) error {
	sd := s.loadFile(id)
	if sd == nil {
		return fmt.Errorf("session %s not found", id)
	}

	s.chat.mu.Lock()
	s.chat.messages = sd.Messages
	s.chat.history = sd.History
	s.chat.model = sd.Model
	s.chat.workDir = sd.WorkDir
	s.chat.mu.Unlock()

	return nil
}

// DeleteSession removes a session file
func (s *SessionService) DeleteSession(id string) error {
	return os.Remove(filepath.Join(s.dir, id+".json"))
}

// NewSession clears the current chat and returns a new session ID
func (s *SessionService) NewSession() string {
	s.chat.ClearHistory()
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func (s *SessionService) loadFile(id string) *SessionData {
	data, err := os.ReadFile(filepath.Join(s.dir, id+".json"))
	if err != nil {
		return nil
	}
	var sd SessionData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil
	}
	return &sd
}
