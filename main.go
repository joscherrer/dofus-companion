package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/system"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/joscherrer/dofus-manager/internal/ui"
	"github.com/joscherrer/dofus-manager/internal/window"

	"gioui.org/widget"
	"gioui.org/widget/material"

	"golang.design/x/hotkey"
)

var (
	clientList      = &ui.ClientList{}
	windowList      []*ClientRow
	windowListMutex = sync.Mutex{}
	windowMap       = make(map[int32]*ClientRow)
	focusIdx        = 0
)

type ClientRow struct {
	w                *window.Window
	Label            *ui.Clickable
	DragHandle       *ui.Clickable
	Draggable        *ui.Draggable
	currentLabelText string
}

func (c *ClientRow) SetWindow(w *window.Window) {
	c.w = w
}

func (c *ClientRow) LabelText() (s string, err error) {
	winTitle, _ := c.w.GetWindowText()
	s, _, _ = strings.Cut(winTitle, " - ")
	c.currentLabelText = s
	return
}

func (c *ClientRow) LabelTextChanged() bool {
	curr := strings.Clone(c.currentLabelText)
	next, err := c.LabelText()
	return err != nil || curr != next
}

func WatchWindows(w *app.Window) {
	for {
		windowListMutex.Lock()
		changed := false
		_windowList, _ := window.GetDofusWindows()
		_windowMap := make(map[int32]*ClientRow)
		for _, w := range _windowList {
			_windowMap[w.Pid()] = &ClientRow{
				w:          &w,
				Label:      &ui.Clickable{},
				DragHandle: &ui.Clickable{},
				Draggable:  &ui.Draggable{Type: "ClientRow"},
			}
		}

		dead := make([]int32, 0)

		// Check for dead windows
		for pid := range windowMap {
			if _, ok := _windowMap[pid]; !ok {
				dead = append(dead, pid)
			}
		}

		// Flag windowList as changed
		if len(dead) > 0 {
			changed = true
		}

		// Remove dead windows from windowMap
		for _, pid := range dead {
			delete(windowMap, pid)
		}

		// Pack windowList, keeping only alive windows
		newWindowList := make([]*ClientRow, 0)
		for _, w := range windowList {
			if _, ok := _windowMap[w.w.Pid()]; ok {
				newWindowList = append(newWindowList, w)
			}
		}

		// Add new windows to windowMap and windowList
		for pid, row := range _windowMap {
			if _, ok := windowMap[pid]; ok {
				// redraw if one of the windows has changed its title
				changed = changed || windowMap[pid].LabelTextChanged()
				continue
			}
			windowMap[pid] = row
			newWindowList = append(newWindowList, row)
			changed = true
		}

		// Trigger update only if there was any change
		if changed {
			fmt.Println("Updating window list")
			windowList = newWindowList
			w.Invalidate()
		}

		windowListMutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func SetWindowBackground(gtx *layout.Context, theme *material.Theme) {
	macro := op.Record(gtx.Ops)
	rect := image.Rectangle{
		Max: image.Point{
			X: gtx.Constraints.Max.X,
			Y: gtx.Constraints.Max.Y,
		},
	}
	paint.FillShape(gtx.Ops, rgb(0x2d2e32), clip.Rect(rect).Op())
	background := macro.Stop()

	background.Add(gtx.Ops)
}

func MakeRows(theme *material.Theme) (rows []*ui.ClientRowStyle) {
	for _, row := range windowList {
		// fmt.Println(i, "b: ", row)
		name, err := row.LabelText()
		if err != nil {
			continue
		}
		_r := &ui.ClientRowStyle{
			Name:       name,
			DragHandle: row.DragHandle,
			Label:      row.Label,
			Draggable:  (*row).Draggable,
			Theme:      theme,
		}
		rows = append(rows, _r)
		// fmt.Println(i, "a: ", _r)
	}
	return
}

func focusNext(w *app.Window) {
	windowList[focusIdx].Label.Active = false
	focusIdx++
	if focusIdx == len(windowList) {
		focusIdx = 0
	}
	windowList[focusIdx].w.UIABringToFront()
	windowList[focusIdx].Label.Active = true
	w.Invalidate()
}

func focusPrev(w *app.Window) {
	windowList[focusIdx].Label.Active = false
	focusIdx--
	if focusIdx < 0 {
		focusIdx = len(windowList) - 1
	}
	windowList[focusIdx].w.UIABringToFront()
	windowList[focusIdx].Label.Active = true
	w.Invalidate()
}

func nextHk(w *app.Window) {
	next := hotkey.New([]hotkey.Modifier{}, hotkey.KeyF3)
	if err := next.Register(); err != nil {
		panic("hotkey registration failed")
	}

	for range next.Keydown() {
		focusNext(w)
	}
}

func prevHk(w *app.Window) {
	prev := hotkey.New([]hotkey.Modifier{}, hotkey.KeyF2)
	if err := prev.Register(); err != nil {
		panic("hotkey registration failed")
	}

	for range prev.Keydown() {
		focusPrev(w)
	}
}

func main() {
	window := new(app.Window)
	go func() {
		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	go WatchWindows(window)
	app.Main()
}

func ColorBox(gtx layout.Context, size image.Point, color color.NRGBA) layout.Dimensions {
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	event.Op(gtx.Ops, &drop)
	return layout.Dimensions{Size: size}
}

var drop int

func run(window *app.Window) error {
	go nextHk(window)
	go prevHk(window)
	window.Option(app.Title("Dofus Manager"))
	window.Option(app.Decorated(false))
	theme := material.NewTheme()
	theme.Palette = defaultPalette

	var ops op.Ops
	var settings ui.Clickable
	var prev ui.Clickable
	var next ui.Clickable
	actions := system.ActionClose | system.ActionMaximize | system.ActionMinimize
	decorations := material.Decorations(theme, &widget.Decorations{}, actions, "Dofus Manager")

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			SetWindowBackground(&gtx, theme)

			window.Perform(decorations.Decorations.Update(gtx))
			deco := func(gtx layout.Context) layout.Dimensions {
				return decorations.Layout(gtx)
			}

			if prev.Clicked(gtx) {
				fmt.Println("Clicked")
			}

			children := make([]layout.FlexChild, 0)
			children = append(children, layout.Rigid(deco))
			children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				size := gtx.Constraints.Constrain(image.Point{})
				rect := clip.Rect{Max: size}.Push(gtx.Ops)
				event.Op(gtx.Ops, &drop)
				rect.Pop()
				return layout.Dimensions{Size: size}
			}))
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return clientList.Layout(gtx, MakeRows(theme)...)
			}))

			for i, row := range windowList {
				if row.Label.Clicked(gtx) {
					fmt.Printf("Bring %s to front\n", row.w.Title())
					go func() {
						windowList[focusIdx].Label.Active = false
						row.w.UIABringToFront()
						row.Label.Active = true
						focusIdx = i
					}()
				}
			}

			if next.Clicked(gtx) {
				focusNext(window)
			}

			if prev.Clicked(gtx) {
				focusPrev(window)
			}

			menu := ui.MenuStyle{
				Theme:    theme,
				Prev:     &prev,
				Next:     &next,
				Settings: &settings,
			}

			menuChild := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return menu.Layout(gtx)
			})
			children = append(children, menuChild)

			flex := layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}

			flex.Layout(gtx, children...)

			for {
				ev, ok := gtx.Event(transfer.TargetFilter{Target: &drop, Type: "ClientRow"})
				if !ok {
					break
				}
				switch e := ev.(type) {
				case transfer.DataEvent:
					data := e.Open()
					defer data.Close()
					content, _ := io.ReadAll(data)
					fmt.Println(string(content))
				}
			}
			// Pass the drawing operations to the GPU.
			e.Frame(gtx.Ops)
		}
	}
}
