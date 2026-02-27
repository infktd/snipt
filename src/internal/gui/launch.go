package gui

import (
	_ "embed"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/infktd/snipt/src/frontend"
	"github.com/infktd/snipt/src/internal/db"
)

//go:embed tray_icon.png
var trayIcon []byte

// LaunchGUI starts the Wails v3 application with manage + find palette windows
// and a system tray. One process, one dock icon, two windows.
func LaunchGUI(store *db.Store, version string) error {
	service := NewSnippetService(store, version)

	app := application.New(application.Options{
		Name: "snipt",
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(frontend.Assets),
		},
	})

	// ── Manage Window ──────────────────────────────────────
	manageWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "manage",
		Title:            "snipt",
		Width:            1100,
		Height:           700,
		MinWidth:         800,
		MinHeight:        500,
		BackgroundColour: application.NewRGBA(24, 25, 38, 255), // #181926 crust
		URL:              "/",
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBar{
				AppearsTransparent: true,
				Hide:               false,
				HideTitle:          true,
				FullSizeContent:    true,
				UseToolbar:         false,
			},
			InvisibleTitleBarHeight: 40,
		},
	})

	// Hide on close instead of quitting — hook runs before the default close handler.
	manageWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		manageWindow.Hide()
		e.Cancel()
	})

	// ── Find Palette Window ────────────────────────────────
	findWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "find",
		Title:            "snipt find",
		Width:            680,
		Height:           420,
		MaxHeight:        500,
		MinWidth:         500,
		Frameless:        true,
		Hidden:           true,
		AlwaysOnTop:      true,
		HideOnFocusLost:  true,
		HideOnEscape:     true,
		BackgroundColour: application.NewRGBA(0, 0, 0, 0),
		URL:              "/?mode=find",
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTransparent,
			TitleBar: application.MacTitleBar{
				AppearsTransparent: true,
				Hide:               true,
				HideTitle:          true,
				FullSizeContent:    true,
			},
		},
	})

	// When frontend signals "find-done" (user selected snippet), hide the palette.
	app.Event.On("find-done", func(e *application.CustomEvent) {
		findWindow.Hide()
	})

	// Dynamic resize: frontend emits desired height based on result count.
	// Only resize height — no Center() here to avoid stealing focus from the input.
	app.Event.On("resize-find", func(e *application.CustomEvent) {
		if data, ok := e.Data.(map[string]interface{}); ok {
			if h, ok := data["height"].(float64); ok {
				height := int(h)
				if height < 100 {
					height = 100
				}
				if height > 500 {
					height = 500
				}
				findWindow.SetSize(680, height)
			}
		}
	})

	// ── System Tray ────────────────────────────────────────
	tray := app.SystemTray.New()
	tray.SetTemplateIcon(trayIcon)

	menu := app.NewMenu()
	menu.Add("Find").OnClick(func(ctx *application.Context) {
		showFindPalette(app, findWindow)
	})
	menu.Add("Manage").OnClick(func(ctx *application.Context) {
		manageWindow.Show()
		manageWindow.Focus()
	})
	menu.AddSeparator()
	menu.Add("Settings").OnClick(func(ctx *application.Context) {
		manageWindow.Show()
		manageWindow.Focus()
		app.Event.Emit("open-settings")
	})
	menu.AddSeparator()
	menu.Add("Quit snipt").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	tray.SetMenu(menu)

	// Click tray icon → show find palette centered (Spotlight-like).
	tray.OnClick(func() {
		showFindPalette(app, findWindow)
	})

	// ── Auto-Sync on Launch ──────────────────────────────
	// Fire a background sync shortly after startup so the app is up-to-date.
	service.triggerAutoSync()

	// ── Dock Icon Click ───────────────────────────────────
	// Wails has a built-in ApplicationShouldHandleReopen listener that shows ALL
	// hidden windows. We use a hook (runs before listeners) to cancel it and
	// show only the manage window.
	app.Event.RegisterApplicationEventHook(events.Mac.ApplicationShouldHandleReopen, func(event *application.ApplicationEvent) {
		manageWindow.Show()
		manageWindow.Focus()
		event.Cancel()
	})

	// ── Application Menu ───────────────────────────────────
	appMenu := app.NewMenu()
	sniptMenu := appMenu.AddSubmenu("snipt")

	sniptMenu.Add("About snipt").OnClick(func(ctx *application.Context) {
		d := app.Dialog.Info().
			SetTitle("snipt").
			SetMessage("Version " + version + "\n\nA snippet manager for the command line and beyond.")
		d.AddButton("OK")
		d.Show()
	})

	sniptMenu.AddSeparator()

	sniptMenu.Add("Settings...").
		SetAccelerator("CmdOrCtrl+,").
		OnClick(func(ctx *application.Context) {
			manageWindow.Show()
			manageWindow.Focus()
			app.Event.Emit("open-settings")
		})

	sniptMenu.AddSeparator()

	sniptMenu.Add("Quit snipt").
		SetAccelerator("CmdOrCtrl+q").
		OnClick(func(ctx *application.Context) {
			app.Quit()
		})

	editMenu := appMenu.AddSubmenu("Edit")
	editMenu.AddRole(application.Undo)
	editMenu.AddRole(application.Redo)
	editMenu.AddSeparator()
	editMenu.AddRole(application.Cut)
	editMenu.AddRole(application.Copy)
	editMenu.AddRole(application.Paste)
	editMenu.AddRole(application.PasteAndMatchStyle)
	editMenu.AddSeparator()
	editMenu.AddRole(application.SelectAll)

	app.Menu.Set(appMenu)

	return app.Run()
}

// showFindPalette centers the find palette, shows it, focuses it,
// and emits a reset event to the frontend.
func showFindPalette(app *application.App, w application.Window) {
	w.Center()
	w.Show()
	w.Focus()
	app.Event.Emit("find-opened")
}
