package gui

import (
	"fyne.io/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"

	"github.com/infktd/snipt/src/frontend"
	"github.com/infktd/snipt/src/internal/db"
)

// LaunchGUI starts a Wails window in the given mode ("manage" or "find").
func LaunchGUI(store *db.Store, mode, version string) error {
	app := NewApp(store, mode, version)

	opts := &options.App{
		Title: "snipt",
		AssetServer: &assetserver.Options{
			Assets: frontend.Assets,
		},
		BackgroundColour: &options.RGBA{R: 13, G: 13, B: 20, A: 255},
		OnStartup:        app.Startup,
		Bind:             []interface{}{app},
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
	}

	switch mode {
	case "find":
		opts.Title = "snipt find"
		opts.Width = 680
		opts.Height = 420
		opts.MaxHeight = 500
		opts.MinWidth = 500
		opts.Frameless = true
		opts.AlwaysOnTop = true
		opts.BackgroundColour = &options.RGBA{R: 36, G: 36, B: 53, A: 255}
		opts.Mac.TitleBar.HideTitleBar = true
	default: // "manage"
		opts.Width = 1100
		opts.Height = 700
		opts.MinWidth = 800
		opts.MinHeight = 500
		opts.HideWindowOnClose = true

		// Start system tray before the Wails event loop.
		go setupTray(app)
	}

	err := wails.Run(opts)

	// Tear down the system tray when Wails exits.
	if mode != "find" {
		systray.Quit()
	}

	return err
}
