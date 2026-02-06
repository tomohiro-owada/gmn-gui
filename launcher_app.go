package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/tomohiro-owada/gmn-gui/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LauncherApp manages the launcher-mode application lifecycle
type LauncherApp struct {
	ctx      context.Context
	mode     *service.ModeService
	settings *service.SettingsService
	chat     *service.ChatService  // not actively used, bound for Wails type generation
	mcp      *service.MCPManager   // not actively used, bound for Wails type generation
	session  *service.SessionService
}

// NewLauncherApp creates a new launcher-mode application
func NewLauncherApp() *LauncherApp {
	settings := service.NewSettingsService()
	mcpMgr := service.NewMCPManager(settings)
	chat := service.NewChatService(settings, mcpMgr)
	session := service.NewSessionService(nil) // read-only, no chat service

	mode := service.NewModeService("launcher", "", "")

	return &LauncherApp{
		mode:     mode,
		settings: settings,
		chat:     chat,
		mcp:      mcpMgr,
		session:  session,
	}
}

// startup is called when the Wails app starts
func (a *LauncherApp) startup(ctx context.Context) {
	a.ctx = ctx
	a.settings.SetContext(ctx)
	a.session.SetContext(ctx)

	// Initialize settings (for auth status display)
	if err := a.settings.Initialize(); err != nil {
		fmt.Printf("Settings initialization warning: %v\n", err)
	}
}

// shutdown is called when the app is closing
func (a *LauncherApp) shutdown(ctx context.Context) {
	// nothing to clean up
}

// OpenProject spawns a new gmn-gui process in chat mode
func (a *LauncherApp) OpenProject(dir, sessionID string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable: %w", err)
	}

	args := []string{"--workdir", dir}
	if sessionID != "" {
		args = append(args, "--session", sessionID)
	}

	cmd := exec.Command(exePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start chat window: %w", err)
	}

	// Detach the child process
	go cmd.Wait()

	return nil
}

// SelectDirectory opens a native directory picker dialog
func (a *LauncherApp) SelectDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Project Directory",
	})
	if err != nil {
		return ""
	}
	return dir
}
