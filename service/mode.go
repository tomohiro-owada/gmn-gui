package service

// ModeService exposes the current app mode to the frontend
type ModeService struct {
	mode      string
	workDir   string
	sessionID string
}

// NewModeService creates a new mode service
func NewModeService(mode, workDir, sessionID string) *ModeService {
	return &ModeService{mode: mode, workDir: workDir, sessionID: sessionID}
}

// GetMode returns "launcher" or "chat"
func (m *ModeService) GetMode() string { return m.mode }

// GetWorkDir returns the working directory (chat mode only)
func (m *ModeService) GetWorkDir() string { return m.workDir }

// GetSessionID returns the current session ID (chat mode only)
func (m *ModeService) GetSessionID() string { return m.sessionID }

// SetSessionID updates the session ID (called when a new session is created)
func (m *ModeService) SetSessionID(id string) { m.sessionID = id }
