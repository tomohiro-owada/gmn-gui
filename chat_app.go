package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tomohiro-owada/gmn-gui/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ChatApp manages the chat-mode application lifecycle
type ChatApp struct {
	ctx       context.Context
	mode      *service.ModeService
	settings  *service.SettingsService
	chat      *service.ChatService
	mcp       *service.MCPManager
	session   *service.SessionService
	workDir   string
	sessionID string
}

// NewChatApp creates a new chat-mode application
func NewChatApp(workDir, sessionID string) *ChatApp {
	// Change to the working directory so config.Load() picks up project-local settings
	// and MCP servers inherit the correct cwd
	if workDir != "" {
		os.Chdir(workDir)
	}

	settings := service.NewSettingsService()
	mcpMgr := service.NewMCPManager(settings)
	chat := service.NewChatService(settings, mcpMgr)
	session := service.NewSessionService(chat)

	mode := service.NewModeService("chat", workDir, sessionID)

	return &ChatApp{
		mode:      mode,
		settings:  settings,
		chat:      chat,
		mcp:       mcpMgr,
		session:   session,
		workDir:   workDir,
		sessionID: sessionID,
	}
}

// startup is called when the Wails app starts
func (a *ChatApp) startup(ctx context.Context) {
	a.ctx = ctx
	a.settings.SetContext(ctx)
	a.chat.SetContext(ctx)
	a.mcp.SetContext(ctx)
	a.session.SetContext(ctx)

	// Initialize settings and auth
	if err := a.settings.Initialize(); err != nil {
		fmt.Printf("Settings initialization warning: %v\n", err)
	}

	// Restore session or create new one
	if a.sessionID != "" {
		if err := a.session.LoadSession(a.sessionID); err != nil {
			fmt.Printf("Failed to load session %s: %v\n", a.sessionID, err)
			a.sessionID = a.session.NewSessionForDir(a.workDir)
			a.mode.SetSessionID(a.sessionID)
		}
	} else {
		a.sessionID = a.session.NewSessionForDir(a.workDir)
		a.mode.SetSessionID(a.sessionID)
	}

	// Always ensure workDir matches this window's directory (set AFTER session load)
	if a.workDir != "" {
		a.chat.SetWorkDir(a.workDir)
	}

	// Auto-connect configured MCP servers
	go a.mcp.ConnectAll()
}

// shutdown is called when the app is closing
func (a *ChatApp) shutdown(ctx context.Context) {
	// Auto-save current session
	if a.sessionID != "" {
		if err := a.session.SaveCurrentSession(a.sessionID); err != nil {
			fmt.Printf("Failed to save session on shutdown: %v\n", err)
		}
	}
	a.mcp.DisconnectAll()
}

// OpenExternal opens a URL in the default browser or a file path in the default app
func (a *ChatApp) OpenExternal(target string) error {
	// If it looks like a relative path, resolve against workDir
	if !filepath.IsAbs(target) && !isURL(target) {
		target = filepath.Join(a.workDir, target)
	}
	return exec.Command("open", target).Start()
}

func isURL(s string) bool {
	return len(s) > 8 && (s[:7] == "http://" || s[:8] == "https://")
}

// SelectDirectory opens a native directory picker dialog
func (a *ChatApp) SelectDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Working Directory",
	})
	if err != nil {
		return ""
	}
	return dir
}
