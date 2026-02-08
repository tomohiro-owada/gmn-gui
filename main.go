package main

import (
	"embed"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Check if running in Wails dev mode by looking for Wails-specific flags
	// In dev mode, Wails adds its own flags like -assetdir, -loglevel, etc.
	isDevMode := false
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-assetdir") || strings.HasPrefix(arg, "-loglevel") {
			isDevMode = true
			break
		}
	}

	// Only parse custom flags if not in dev mode
	var workDirVal, sessionIDVal string
	if !isDevMode {
		workDir := flag.String("workdir", "", "Working directory for chat mode")
		sessionID := flag.String("session", "", "Session ID to restore")
		flag.Parse()
		workDirVal = *workDir
		sessionIDVal = *sessionID
	}

	if workDirVal != "" {
		runChatMode(workDirVal, sessionIDVal)
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
