package gui

/*
#include <stdlib.h>
#include "tray_darwin.h"
*/
import "C"

import (
	"unsafe"

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
		showManageWindow(trayApp)
	}
}

//export goMenuSettings
func goMenuSettings() {
	if trayApp != nil {
		showSettings(trayApp)
	}
}

//export goMenuQuit
func goMenuQuit() {
	if trayApp != nil {
		wailsRuntime.Quit(trayApp.ctx)
	}
}

func setupTray(app *App) {
	trayApp = app
	ptr := unsafe.Pointer(&trayIconBytes[0])
	C.setupNativeTray(ptr, C.int(len(trayIconBytes)))
}

func teardownTray() {
	C.teardownNativeTray()
	trayApp = nil
}
