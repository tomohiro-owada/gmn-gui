package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/tomohiro-owada/gmn-gui/internal/api"
	"github.com/tomohiro-owada/gmn-gui/internal/auth"
	"github.com/tomohiro-owada/gmn-gui/internal/config"
)

// SettingsService manages configuration and authentication state
type SettingsService struct {
	ctx       context.Context
	mu        sync.RWMutex
	config    *config.Config
	authMgr   *auth.Manager
	projectID string
	model     string
}

// NewSettingsService creates a new settings service
func NewSettingsService() *SettingsService {
	return &SettingsService{
		model: "gemini-2.5-flash",
	}
}

// SetContext sets the Wails runtime context
func (s *SettingsService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// Initialize loads config and checks auth
func (s *SettingsService) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	s.config = cfg

	mgr, err := auth.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create auth manager: %w", err)
	}
	s.authMgr = mgr

	// Try to load cached project ID
	state, err := config.LoadCachedState()
	if err == nil && state.ProjectID != "" {
		s.projectID = state.ProjectID
	}

	return nil
}

// AuthStatus represents the authentication state for the frontend
type AuthStatus struct {
	Authenticated bool   `json:"authenticated"`
	ProjectID     string `json:"projectId"`
	Error         string `json:"error,omitempty"`
}

// GetAuthStatus checks the current authentication status
func (s *SettingsService) GetAuthStatus() AuthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.authMgr == nil {
		return AuthStatus{Error: "auth manager not initialized"}
	}

	creds, err := s.authMgr.LoadCredentials()
	if err != nil {
		return AuthStatus{Error: err.Error()}
	}

	if creds.IsExpired() {
		_, err := s.authMgr.RefreshToken(creds)
		if err != nil {
			return AuthStatus{Error: "token expired: " + err.Error()}
		}
	}

	return AuthStatus{
		Authenticated: true,
		ProjectID:     s.projectID,
	}
}

// EnsureAuth returns an authenticated API client, refreshing tokens if needed
func (s *SettingsService) EnsureAuth(ctx context.Context) (*api.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.authMgr == nil {
		return nil, fmt.Errorf("not initialized")
	}

	creds, err := s.authMgr.LoadCredentials()
	if err != nil {
		return nil, err
	}

	if creds.IsExpired() {
		creds, err = s.authMgr.RefreshToken(creds)
		if err != nil {
			return nil, err
		}
	}

	httpClient := s.authMgr.HTTPClient(creds)
	client := api.NewClient(httpClient)

	// LoadCodeAssist if no project ID cached
	if s.projectID == "" {
		resp, err := client.LoadCodeAssist(ctx)
		if err != nil {
			return nil, fmt.Errorf("LoadCodeAssist failed: %w", err)
		}
		s.projectID = resp.CloudAICompanionProject
		if s.projectID != "" {
			_ = config.SaveCachedState(&config.CachedState{ProjectID: s.projectID})
		}
	}

	return client, nil
}

// GetProjectID returns the cached project ID
func (s *SettingsService) GetProjectID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projectID
}

// GetConfig returns the current config
func (s *SettingsService) GetConfig() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// ReloadConfig reloads configuration from disk
func (s *SettingsService) ReloadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	s.config = cfg
	return nil
}

// GetModel returns the current model
func (s *SettingsService) GetModel() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.model
}

// SetModel sets the model
func (s *SettingsService) SetModel(model string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.model = model
}

// AvailableModels returns the list of available models
func (s *SettingsService) AvailableModels() []string {
	return []string{
		"gemini-2.5-pro",
		"gemini-2.5-flash",
		"gemini-2.0-flash",
	}
}
