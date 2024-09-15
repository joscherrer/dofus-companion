package window

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/joscherrer/dofus-manager/internal/config"
	"golang.org/x/sys/windows"
)

type manager struct {
	hwnds           []windows.Handle
	sortedHwnds     []windows.Handle
	sortedWindows   []*Window
	windows         map[windows.Handle]*Window
	hwndsLock       sync.Mutex
	enumCallbackPtr uintptr
	changedChan     chan bool
	activeHwnd      windows.Handle
	SetKey          string
	SetKeyUpdated   bool
}

var (
	m    *manager
	once sync.Once
)

const (
	emptyTitle = "..."
	procFilter = "Dofus.exe"
)

func (m *manager) enumCallback(h windows.Handle, _ uintptr) uintptr {
	w := NewWindow(h)
	if w.ProcessName() != procFilter {
		return 1
	}
	if !windows.IsWindowVisible(windows.HWND(w.hwnd)) {
		return 1
	}
	m.hwnds = append(m.hwnds, h)

	if _, ok := m.windows[h]; !ok {
		m.changedChan <- true
		m.windows[h] = &w
		m.windows[h].Title()
		return 1
	}

	if m.windows[h].TitleChanged() {
		m.changedChan <- true
	}
	return 1
}

func (m *manager) enumWindows() {
	windows.EnumWindows(m.enumCallbackPtr, nil)
	if len(m.hwnds) != len(m.sortedHwnds) {
		m.changedChan <- true
	}
	close(m.changedChan)
}

func (m *manager) UpdateWindows() (changed bool) {
	m.hwndsLock.Lock()
	defer m.hwndsLock.Unlock()
	changed = false
	m.hwnds = make([]windows.Handle, 0)
	m.changedChan = make(chan bool)
	go m.enumWindows()
	for c := range m.changedChan {
		changed = c || changed
	}

	if !changed {
		return
	}

	fmt.Println("Windows changed")
	m.restoreOrder()
	m.sortHwnds()

	for i, hwnd := range m.sortedHwnds {
		m.windows[hwnd].order = i
		m.sortedWindows = append(m.sortedWindows, m.windows[hwnd])
	}
	if m.activeHwnd == 0 {
		m.activeHwnd = m.sortedHwnds[0]
	}
	// m.saveOrder()
	return
}

func (m *manager) sortHwnds() {
	m.sortedHwnds = make([]windows.Handle, len(m.hwnds))
	copy(m.sortedHwnds, m.hwnds)
	slices.SortStableFunc(m.sortedHwnds, func(i, j windows.Handle) int {
		if m.windows[i].order < m.windows[j].order {
			return -1
		} else if m.windows[i].order > m.windows[j].order {
			return 1
		} else {
			return 0
		}
	})
}

func (m *manager) restoreOrder() {
	fmt.Println("restoreOrder")
	orderKey := m.GetWindowSetKey()

	configWindows := config.GetConfig().Windows
	mapOrder := make(map[string]config.WindowConfig)

	m.SetKey = orderKey
	m.SetKeyUpdated = true
	if _, ok := configWindows[orderKey]; !ok {
		return
	}
	fmt.Printf("Restoring order: %s\n", orderKey)
	for _, c := range configWindows[orderKey] {
		mapOrder[c.Name] = c
	}

	for _, hwnd := range m.hwnds {
		label := m.GetWindowLabel(hwnd)
		if w, ok := mapOrder[label]; ok {
			m.windows[hwnd].order = w.Order
		} else {
			m.windows[hwnd].order = 0
		}
	}
}

func (m *manager) saveOrder(setKey string) {
	c := config.GetConfig()
	orderKey := m.GetWindowSetKey()

	winMap := make(map[string]windows.Handle)
	for _, hwnd := range m.hwnds {
		winMap[m.GetWindowLabel(hwnd)] = hwnd
	}

	configWindows := make([]config.WindowConfig, 0)
	for _, hwnd := range m.sortedHwnds {
		win := m.windows[hwnd]
		label := m.GetWindowLabel(hwnd)
		if label == emptyTitle {
			continue
		}
		configWindows = append(configWindows, config.WindowConfig{
			Name:  label,
			Order: win.order,
		})
	}

	if len(configWindows) == 0 {
		return
	}

	if orderKey != "" && orderKey != setKey {
		fmt.Println("Renaming window set")
		delete(c.Windows, orderKey)
	}
	m.SetKey = setKey
	c.Windows[m.SetKey] = configWindows
	config.SetConfig(c)
	// confHandler := config.GetConfigHandler()
	// confHandler.Set("windows", c.Windows)
	// confHandler.WriteConfig()
}

func (m *manager) GetSortedHwnds() []windows.Handle {
	return m.sortedHwnds
}

func (m *manager) GetWindow(h windows.Handle) *Window {
	return m.windows[h]
}

func (m *manager) GetWindowLabel(h windows.Handle) (s string) {
	winTitle := m.GetWindow(h).Title()
	s, _, found := strings.Cut(winTitle, " - ")
	if !found {
		s = emptyTitle
	}
	if strings.Contains(s, "Dofus") {
		s = emptyTitle
	}
	return
}

func (m *manager) GetWindowSetKey() string {
	fmt.Println("GetWindowSetKey")
	orderKeySlice := make([]string, 0)
	for _, hwnd := range m.hwnds {
		label := m.GetWindowLabel(hwnd)
		if label == emptyTitle {
			continue
		}
		orderKeySlice = append(orderKeySlice, m.GetWindowLabel(hwnd))
	}
	slices.Sort(orderKeySlice)

	c := config.GetConfig()
	for k, v := range c.Windows {
		configOrder := make([]string, 0)
		for _, w := range v {
			configOrder = append(configOrder, w.Name)
		}
		slices.Sort(configOrder)
		if slices.Equal(orderKeySlice, configOrder) {
			return k
		}
	}

	return ""
}

func (m *manager) MoveWindow(h windows.Handle, offset int) {
	m.hwndsLock.Lock()
	defer m.hwndsLock.Unlock()
	offset = offset % len(m.sortedHwnds)
	initialPos := m.windows[h].order
	newPos := initialPos + offset
	for _, hwnd := range m.sortedHwnds {
		win := m.windows[hwnd]
		switch {
		case offset > 0 && win.order > initialPos && win.order <= newPos:
			win.order--
		case offset < 0 && win.order < initialPos && win.order >= newPos:
			win.order++
		case win.order == initialPos:
			win.order = newPos
		}
	}
	m.sortHwnds()
	m.saveOrder(m.GetWindowSetKey())
}

func (m *manager) BringToFront(h windows.Handle) {
	m.GetWindow(h).BringToFront()
	m.activeHwnd = h
}

func (m *manager) FocusNext() {
	if len(m.sortedHwnds) == 0 {
		return
	}
	next := (m.GetWindow(m.activeHwnd).order + 1) % len(m.sortedHwnds)
	m.BringToFront(m.sortedHwnds[next])
	m.activeHwnd = m.sortedHwnds[next]
}

func (m *manager) FocusPrev() {
	if len(m.sortedHwnds) == 0 {
		return
	}
	prev := (m.GetWindow(m.activeHwnd).order - 1 + len(m.sortedHwnds)) % len(m.sortedHwnds)
	m.BringToFront(m.sortedHwnds[prev])
	m.activeHwnd = m.sortedHwnds[prev]
}

func (m *manager) ActiveHwnd() windows.Handle {
	return m.activeHwnd
}

func (m *manager) SaveOrder(setKey string) {
	if setKey == "" {
		fmt.Println("No set key provided")
		return
	}
	m.hwndsLock.Lock()
	m.saveOrder(setKey)
	m.restoreOrder()
	m.hwndsLock.Unlock()
}

func GetManager() *manager {
	once.Do(func() {
		m = &manager{
			hwnds:       make([]windows.Handle, 0),
			sortedHwnds: make([]windows.Handle, 0),
			windows:     make(map[windows.Handle]*Window),
			hwndsLock:   sync.Mutex{},
			changedChan: make(chan bool),
			activeHwnd:  0,
		}
		m.enumCallbackPtr = syscall.NewCallback(m.enumCallback)
		m.UpdateWindows()
		ch := config.SubscribeToChanged()
		go func() {
			for range ch {
				m.restoreOrder()
			}
		}()
	})
	return m
}
