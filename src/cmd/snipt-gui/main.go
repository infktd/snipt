package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/infktd/snipt/src/internal/config"
	"github.com/infktd/snipt/src/internal/db"
	"github.com/infktd/snipt/src/internal/gui"
)

//go:embed all:../../frontend/dist
var assets embed.FS

func main() {
	dbPath := config.DBPath("")
	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer store.Close()

	app := gui.NewApp(store)

	err = wails.Run(&options.App{
		Title:     "snipt",
		Width:     1100,
		Height:    700,
		MinWidth:  800,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 13, G: 13, B: 20, A: 255},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                 true,
				HideTitleBar:              false,
				FullSizeContent:           true,
				UseToolbar:                false,
			},
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		log.Fatalf("wails: %v", err)
	}
}
