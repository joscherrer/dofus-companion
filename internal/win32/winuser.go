package win32

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type (
	HRESULT int32
	HKL     syscall.Handle
)

const (
	HWND_BOTTOM        = windows.Handle(1)
	HWND_TOP           = windows.Handle(0)
	HWND_NOTOPMOST     = ^windows.Handle(1) // -2
	HWND_TOPMOST       = ^windows.Handle(0) // -1
	SWP_NOSIZE         = 0x0001
	SWP_NOMOVE         = 0x0002
	SWP_SHOWWINDOW     = 0x0040
	MAPVK_VK_TO_VSC    = 0
	MAPVK_VSC_TO_VK    = 1
	MAPVK_VK_TO_CHAR   = 2
	MAPVK_VSC_TO_VK_EX = 3
	MAPVK_VK_TO_VSC_EX = 4
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
	isIconic            = user32.NewProc("IsIconic")
	showWindow          = user32.NewProc("ShowWindow")
	vkKeyScanExA        = user32.NewProc("VkKeyScanExA")
	getKeyboardLayout   = user32.NewProc("GetKeyboardLayout")
	mapVirtualKeyA      = user32.NewProc("MapVirtualKeyA")
	mapVirtualKeyExA    = user32.NewProc("MapVirtualKeyExA")
	getKeyNameText      = user32.NewProc("GetKeyNameTextA")
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

func IsIconic(hwnd windows.Handle) bool {
	ret, _, _ := isIconic.Call(uintptr(hwnd))
	return ret != 0
}

func ShowWindow(hwnd windows.Handle, nCmdShow int) bool {
	ret, _, _ := showWindow.Call(uintptr(hwnd), uintptr(nCmdShow))
	return ret != 0
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
	if length == 0 {
		return "", err
	}
	s = syscall.UTF16ToString(buf[:length])
	err = nil
	return
}

func VkKeyScanExA(ch byte, dHkl windows.Handle) (int, int) {
	ret, _, _ := vkKeyScanExA.Call(uintptr(ch), uintptr(dHkl))
	return int(ret & 0xFF), int(ret >> 8)
}

func MapVirtualKeyA(uCode, uMapType uint32) uint32 {
	ret, _, _ := mapVirtualKeyA.Call(uintptr(uCode), uintptr(uMapType))
	return uint32(ret)
}

func MapVirtualKeyExA(uCode, uMapType uint32, dHkl windows.Handle) uint32 {
	ret, _, _ := mapVirtualKeyExA.Call(uintptr(uCode), uintptr(uMapType), uintptr(dHkl))
	return uint32(ret)
}

func int32ToUintptr(v int32) uintptr {
	if v < 0 {
		return uintptr(uint32(1<<32-1) - uint32(-v) + 1)
	}
	return uintptr(v)
}

// https://msdn.microsoft.com/en-us/library/ms646300.aspx
func GetKeyNameText(lParam int32) (string, error) {
	buf := make([]uint16, 256)
	r1, _, err := getKeyNameText.Call(int32ToUintptr(lParam), uintptr(unsafe.Pointer(&buf[0])), 256)
	if r1 == 0 {
		fmt.Printf("buf: %v\n", buf)
		return "", err
	}
	t := syscall.UTF16ToString(buf[:r1])
	return t, err
}
