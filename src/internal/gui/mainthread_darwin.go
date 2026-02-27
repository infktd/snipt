package gui

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include "mainthread_darwin.h"
*/
import "C"

var mainThreadFunc func()

//export goMainThreadCallback
func goMainThreadCallback() {
	if fn := mainThreadFunc; fn != nil {
		fn()
	}
}

// runOnMainThread dispatches fn to execute on the macOS main thread.
func runOnMainThread(fn func()) {
	mainThreadFunc = fn
	C.dispatchOnMainQueue()
}
