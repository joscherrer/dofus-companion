package win32

import (
	"log"
	"os"
	"regexp"
	"syscall"
	"unsafe"

	ole "github.com/go-ole/go-ole"
	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/windows"
)

type HRESULT int32

const (
	HWND_BOTTOM    = windows.HWND(1)
	HWND_TOP       = windows.HWND(0)
	HWND_NOTOPMOST = ^windows.HWND(1) // -2
	HWND_TOPMOST   = ^windows.HWND(0) // -1
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
	getWindowTextLength = user32.NewProc("GetWindowTextLengthW")
	coInitializeEx      = user32.NewProc("CoInitializeEx")
	coCreateInstance    = user32.NewProc("CoCreateInstance")
)

type Window struct {
	hwnd windows.Handle
	p    *process.Process
}

func (w *Window) GetWindowTextLength() int {
	ret, _, _ := getWindowTextLength.Call(uintptr(windows.HWND(w.hwnd)))
	return int(ret)
}

func (w *Window) Pid() int32 {
	return w.p.Pid
}

func (w *Window) GetWindowText() (s string, err error) {
	textLen := w.GetWindowTextLength() + 1
	buf := make([]uint16, textLen)
	len, _, _ := getWindowText.Call(uintptr(windows.HWND(w.hwnd)), uintptr(unsafe.Pointer(&buf[0])), uintptr(textLen))
	s = syscall.UTF16ToString(buf[:len])
	return
}

func (w *Window) GetWindowThreadProcessId() (pid uint32, tid uint32, err error) {
	tid, err = windows.GetWindowThreadProcessId(windows.HWND(w.hwnd), &pid)
	return
}

func (w *Window) Title() string {
	title, _ := w.GetWindowText()
	return title
}

func NewWindow(h windows.Handle) (w Window) {
	w.hwnd = h
	pid, _, err := w.GetWindowThreadProcessId()
	if err != nil {
		log.Println("Error getting process id for window")
	}
	w.p = &process.Process{Pid: int32(pid)}
	return
}

func ListWindows() ([]Window, error) {
	hwnds := make([]Window, 0)

	cb := windows.NewCallback(func(h windows.Handle, p unsafe.Pointer) uintptr {
		w := NewWindow(h)
		if !windows.IsWindowVisible(windows.HWND(w.hwnd)) {
			return 1
		}
		hwnds = append(hwnds, w)
		return 1
	})
	windows.EnumWindows(cb, nil)
	if hwnds == nil {
		log.Println("No windows found")
	}
	return hwnds, nil
}

func GetDofusWindows() (wins []Window, err error) {
	wl, _ := ListWindows()
	for _, w := range wl {
		name, _ := w.p.Name()
		if name != "Dofus.exe" {
			continue
		}
		wins = append(wins, w)
	}
	return
}

func FilterWindows(pattern string) ([]Window, error) {
	windows, err := ListWindows()
	if err != nil {
		return nil, err
	}
	filtered := make([]Window, 0)
	for _, w := range windows {
		if (int)(w.p.Pid) == os.Getpid() {
			continue
		}
		text, _ := w.GetWindowText()
		matched, _ := regexp.MatchString(pattern, text)
		if matched {
			filtered = append(filtered, w)
		}
	}
	return filtered, nil
}

func (w *Window) UIABringToFront() {
	ole.CoInitialize(0)
	uia, _ := NewUIAutomation()
	el, _ := uia.ElementFromHandle(windows.HWND(w.hwnd))
	el.SetFocus()
	el.Release()
	uia.Release()
	ole.CoUninitialize()
}

// https://stackoverflow.com/questions/916259/win32-bring-a-window-to-top
func (w *Window) BringToFront() {
	foreground := windows.GetForegroundWindow()
	currentThread := windows.GetCurrentThreadId()
	currentThreadPid, _ := windows.GetWindowThreadProcessId(foreground, nil)
	AttachThreadInput(currentThreadPid, currentThread, true)
	SetWindowPos(windows.HWND(w.hwnd), HWND_TOPMOST, 0, 0, 0, 0, SWP_NOSIZE|SWP_NOMOVE)
	SetWindowPos(windows.HWND(w.hwnd), HWND_NOTOPMOST, 0, 0, 0, 0, SWP_SHOWWINDOW|SWP_NOSIZE|SWP_NOMOVE)
	SetForegroundWindow(windows.HWND(w.hwnd))
	SetFocus(windows.HWND(w.hwnd))
	SetActiveWindow(windows.HWND(w.hwnd))
	AttachThreadInput(currentThreadPid, currentThread, false)
}

func AttachThreadInput(tid uint32, id uint32, attach bool) bool {
	_attach := 0
	if attach {
		_attach = 1
	}
	ret, _, _ := attachThreadInput.Call(uintptr(tid), uintptr(id), uintptr(_attach))
	return ret != 0
}

func SetWindowPos(hwnd windows.HWND, hWndInsertAfter windows.HWND, x, y, cx, cy int, uFlags uint) bool {
	ret, _, _ := setWindowPos.Call(
		uintptr(hwnd), uintptr(hWndInsertAfter), uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(uFlags),
	)
	return ret != 0
}

func SetForegroundWindow(window windows.HWND) bool {
	ret, _, _ := setForegroundWindow.Call(uintptr(window))
	return ret != 0
}

func SetFocus(hwnd windows.HWND) windows.HWND {
	ret, _, _ := psetFocus.Call(uintptr(hwnd))
	return windows.HWND(ret)
}

func SetActiveWindow(hwnd windows.HWND) windows.HWND {
	ret, _, _ := setActiveWindow.Call(uintptr(hwnd))
	return windows.HWND(ret)
}
