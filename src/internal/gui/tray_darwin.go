package gui

/*
#include <stdlib.h>
#include "tray_darwin.h"
*/
import "C"

import (
	"unsafe"

	"github.com/pkg/browser"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var trayApp *App

//export goTrayClicked
func goTrayClicked() {
	launchFindPalette()
}

//export goMenuManage
func goMenuManage() {
	if trayApp != nil {
		go showManageWindow(trayApp)
	}
}

//export goMenuSettings
func goMenuSettings() {
	if trayApp != nil {
		go showSettings(trayApp)
	}
}

//export goMenuCheckForUpdates
func goMenuCheckForUpdates() {
	go browser.OpenURL("https://github.com/infktd/snipt/releases/latest")
}

//export goMenuQuit
func goMenuQuit() {
	if trayApp != nil {
		go wailsRuntime.Quit(trayApp.ctx)
	}
}

func setupTray(app *App) {
	trayApp = app
	ptr := unsafe.Pointer(&trayIconBytes[0])
	ver := C.CString(app.version)
	defer C.free(unsafe.Pointer(ver))
	C.setupNativeTray(ptr, C.int(len(trayIconBytes)), ver)
	C.injectAppMenuItems()
}

func teardownTray() {
	C.teardownNativeTray()
	trayApp = nil
}
