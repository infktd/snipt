package gui

import (
	_ "embed"
	"os"
	"os/exec"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed tray_icon.png
var trayIconBytes []byte

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
