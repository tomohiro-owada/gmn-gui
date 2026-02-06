package main

import (
	"embed"
	"flag"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	workDir := flag.String("workdir", "", "Working directory for chat mode")
	sessionID := flag.String("session", "", "Session ID to restore")
	flag.Parse()

	if *workDir != "" {
		runChatMode(*workDir, *sessionID)
	} else {
		runLauncherMode()
	}
}

func runChatMode(workDir, sessionID string) {
	app := NewChatApp(workDir, sessionID)

	title := filepath.Base(workDir) + " - gmn-gui"

	err := wails.Run(&options.App{
		Title:  title,
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnShutdown:        app.shutdown,
		Bind: []interface{}{
			app,
			app.mode,
			app.settings,
			app.chat,
			app.mcp,
			app.session,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func runLauncherMode() {
	app := NewLauncherApp()

	err := wails.Run(&options.App{
		Title:  "gmn-gui",
		Width:  500,
		Height: 400,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnShutdown:        app.shutdown,
		Bind: []interface{}{
			app,
			app.mode,
			app.settings,
			app.chat,
			app.mcp,
			app.session,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
