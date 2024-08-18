package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"syscall"
	"unsafe"

	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/windows"
)

var (
	user32             = syscall.MustLoadDLL("user32.dll")
	procGetWindowTextW = user32.MustFindProc("GetWindowTextW")
)

type Window struct {
	hwnd windows.Handle
	p    *process.Process
}

func (w *Window) GetWindowText() (s string, err error) {
	b := make([]uint16, 200)
	r0, _, e1 := syscall.SyscallN(procGetWindowTextW.Addr(), uintptr(w.hwnd), uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)))
	blen := int32(r0)
	if blen == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	s = syscall.UTF16ToString(b)
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
		fmt.Println(w.GetWindowText())
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
