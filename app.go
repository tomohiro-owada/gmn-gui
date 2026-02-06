package main

import (
	"context"
	"fmt"

	"github.com/tomohiro-owada/gmn-gui/service"
)

// App struct manages the application lifecycle and holds service references
type App struct {
	ctx      context.Context
	settings *service.SettingsService
	chat     *service.ChatService
	mcp      *service.MCPManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	settings := service.NewSettingsService()
	mcpMgr := service.NewMCPManager(settings)
	chat := service.NewChatService(settings, mcpMgr)

	return &App{
		settings: settings,
		chat:     chat,
		mcp:      mcpMgr,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.settings.SetContext(ctx)
	a.chat.SetContext(ctx)
	a.mcp.SetContext(ctx)

	// Initialize settings and auth
	if err := a.settings.Initialize(); err != nil {
		fmt.Printf("Settings initialization warning: %v\n", err)
	}

	// Auto-connect configured MCP servers
	go a.mcp.ConnectAll()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.mcp.DisconnectAll()
}

// Greet returns a greeting for testing wails bindings
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
