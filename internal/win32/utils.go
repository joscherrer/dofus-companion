package win32

import (
	"github.com/go-ole/go-ole"
	"golang.org/x/sys/windows"
)

func UIABringToFront(hwnd windows.Handle) {
	ole.CoInitialize(0)
	uia, _ := NewUIAutomation()
	el, _ := uia.ElementFromHandle(hwnd)
	el.SetFocus()
	el.Release()
	uia.Release()
	ole.CoUninitialize()
}

// https://stackoverflow.com/questions/916259/win32-bring-a-window-to-top
func BringToFront(hwnd windows.Handle) {
	foreground := windows.GetForegroundWindow()
	currentThread := windows.GetCurrentThreadId()
	currentThreadPid, _ := windows.GetWindowThreadProcessId(foreground, nil)
	AttachThreadInput(currentThreadPid, currentThread, true)
	SetWindowPos(hwnd, HWND_TOPMOST, 0, 0, 0, 0, SWP_NOSIZE|SWP_NOMOVE)
	SetWindowPos(hwnd, HWND_NOTOPMOST, 0, 0, 0, 0, SWP_SHOWWINDOW|SWP_NOSIZE|SWP_NOMOVE)
	SetForegroundWindow(hwnd)
	SetFocus(hwnd)
	SetActiveWindow(hwnd)
	AttachThreadInput(currentThreadPid, currentThread, false)
}
