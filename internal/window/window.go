package window

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"syscall"
	"unsafe"

	"github.com/joscherrer/dofus-manager/internal/win32"
	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/windows"
)

type Window struct {
	hwnd  windows.Handle
	p     *process.Process
	order int
	title string
}

var (
	hwnds           []Window = make([]Window, 0)
	enumCallbackPtr uintptr  = syscall.NewCallback(enumCallback)
)

func enumCallback(h windows.Handle, _ unsafe.Pointer) uintptr {
	w := NewWindow(h)
	if !windows.IsWindowVisible(windows.HWND(w.hwnd)) {
		return 1
	}
	hwnds = append(hwnds, w)
	return 1
}

func (w *Window) GetWindowTextLength() (l int, err error) {
	l, err = win32.GetWindowTextLength(w.hwnd)
	return
}

func (w *Window) Handle() windows.Handle {
	return w.hwnd
}

func (w *Window) Pid() int32 {
	return w.p.Pid
}

func (w *Window) GetWindowText() (s string, err error) {
	s, err = win32.GetWindowText(w.hwnd)
	return
}

func (w *Window) GetWindowThreadProcessId() (pid uint32, tid uint32, err error) {
	tid, err = windows.GetWindowThreadProcessId(windows.HWND(w.hwnd), &pid)
	return
}

func (w *Window) UpdateTitle() (next string) {
	next, err := w.GetWindowText()
	if err != nil {
		next = "Closing..."
		fmt.Println(err)
	}
	w.title = next
	return
}

func (w *Window) TitleChanged() (changed bool) {
	curr := strings.Clone(w.title)
	next := w.UpdateTitle()
	return curr != next
}

func (w *Window) Title() string {
	return w.title
}

func (w *Window) ProcessName() string {
	name, _ := w.p.Name()
	return name
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
	windows.EnumWindows(enumCallbackPtr, nil)
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
	win32.UIABringToFront(windows.Handle(w.hwnd))
}

func (w *Window) BringToFront() {
	win32.BringToFront(windows.Handle(w.hwnd))
}
