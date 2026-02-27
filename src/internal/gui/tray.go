package gui

import (
	_ "embed"
	"os"
	"os/exec"
	"time"

	"fyne.io/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed tray_icon.png
var trayIconBytes []byte

// setupTray starts the system tray icon and menu.
// It blocks until systray.Quit() is called.
func setupTray(app *App) {
	systray.Run(func() {
		onTrayReady(app)
	}, func() {
		// onExit — nothing to clean up
	})
}

func onTrayReady(app *App) {
	systray.SetIcon(trayIconBytes)
	systray.SetTooltip("snipt")

	// Title item (disabled, just a label)
	mTitle := systray.AddMenuItem("✂ snipt", "")
	mTitle.Disable()

	systray.AddSeparator()

	mFind := systray.AddMenuItem("Find", "Open find palette")
	mManage := systray.AddMenuItem("Manage", "Show manage window")

	systray.AddSeparator()

	mSettings := systray.AddMenuItem("Settings", "Open settings")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit snipt", "Quit the application")

	go func() {
		for {
			select {
			case <-mFind.ClickedCh:
				launchFindPalette()
			case <-mManage.ClickedCh:
				showManageWindow(app)
			case <-mSettings.ClickedCh:
				showSettings(app)
			case <-mQuit.ClickedCh:
				systray.Quit()
				wailsRuntime.Quit(app.ctx)
				return
			}
		}
	}()
}

// launchFindPalette spawns "snipt find" as a separate process.
func launchFindPalette() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(exe, "find")
	cmd.Start()
}

// showManageWindow brings the manage window to front.
func showManageWindow(app *App) {
	// Wait for Wails context to be available
	for app.ctx == nil {
		time.Sleep(50 * time.Millisecond)
	}
	wailsRuntime.Show(app.ctx)
	wailsRuntime.WindowShow(app.ctx)
}

// showSettings shows the manage window and emits an event to open settings.
func showSettings(app *App) {
	showManageWindow(app)
	wailsRuntime.EventsEmit(app.ctx, "open-settings")
}
