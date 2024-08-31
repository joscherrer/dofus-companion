package win32

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type HRESULT int32

const (
	HWND_BOTTOM    = windows.Handle(1)
	HWND_TOP       = windows.Handle(0)
	HWND_NOTOPMOST = ^windows.Handle(1) // -2
	HWND_TOPMOST   = ^windows.Handle(0) // -1
	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002
	SWP_SHOWWINDOW = 0x0040
)

var (
	user32              = windows.NewLazySystemDLL("user32.dll")
	psetFocus           = user32.NewProc("SetFocus")
	setWindowPos        = user32.NewProc("SetWindowPos")
	getWindowText       = user32.NewProc("GetWindowTextW")
	setActiveWindow     = user32.NewProc("SetActiveWindow")
	attachThreadInput   = user32.NewProc("AttachThreadInput")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
	coInitializeEx      = user32.NewProc("CoInitializeEx")
	coCreateInstance    = user32.NewProc("CoCreateInstance")
	getWindowTextLength = user32.NewProc("GetWindowTextLengthW")
)

func AttachThreadInput(tid uint32, id uint32, attach bool) bool {
	_attach := 0
	if attach {
		_attach = 1
	}
	ret, _, _ := attachThreadInput.Call(uintptr(tid), uintptr(id), uintptr(_attach))
	return ret != 0
}

func SetWindowPos(hwnd windows.Handle, hWndInsertAfter windows.Handle, x, y, cx, cy int, uFlags uint) bool {
	ret, _, _ := setWindowPos.Call(
		uintptr(hwnd), uintptr(hWndInsertAfter), uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(uFlags),
	)
	return ret != 0
}

func SetForegroundWindow(window windows.Handle) bool {
	ret, _, _ := setForegroundWindow.Call(uintptr(window))
	return ret != 0
}

func SetFocus(hwnd windows.Handle) windows.Handle {
	ret, _, _ := psetFocus.Call(uintptr(hwnd))
	return windows.Handle(ret)
}

func SetActiveWindow(hwnd windows.Handle) windows.Handle {
	ret, _, _ := setActiveWindow.Call(uintptr(hwnd))
	return windows.Handle(ret)
}

func GetWindowTextLength(hwnd windows.Handle) (l int, err error) {
	ret, _, err := getWindowTextLength.Call(uintptr(hwnd))
	l = int(ret)
	return
}

func GetWindowText(hwnd windows.Handle) (s string, err error) {
	textLen, _ := GetWindowTextLength(hwnd)
	textLen++
	buf := make([]uint16, textLen)
	length, _, err := getWindowText.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(textLen))
	s = syscall.UTF16ToString(buf[:length])
	return
}
