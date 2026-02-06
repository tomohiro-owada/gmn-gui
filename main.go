package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "gmn-gui",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 14, B: 27, A: 1},
		OnStartup:        app.startup,
		OnShutdown:        app.shutdown,
		Bind: []interface{}{
			app,
			app.settings,
			app.chat,
			app.mcp,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
