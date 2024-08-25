package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/joscherrer/dofus-manager/internal/ui"
	"github.com/joscherrer/dofus-manager/internal/win32"

	"gioui.org/widget"
	"gioui.org/widget/material"

	"golang.design/x/hotkey"
)

var (
	windowList  []win32.Window
	windowList2 []ClientRow
)
var windowMap = make(map[int32]ClientRow)

type ClientRow struct {
	w          *win32.Window
	label      *widget.Clickable
	dragHandle *widget.Clickable
}

func (c *ClientRow) SetWindow(w *win32.Window) {
	c.w = w
}

func WatchWindows(w *app.Window) {
	for {
		_windowList, _ := win32.GetDofusWindows()
		_windowMap := make(map[int32]ClientRow)
		for _, w := range _windowList {
			_windowMap[w.Pid()] = ClientRow{
				w:          &w,
				label:      new(widget.Clickable),
				dragHandle: new(widget.Clickable),
			}
		}

		dead := make([]int32, 0)

		// Check for dead windows
		for pid := range windowMap {
			if _, ok := _windowMap[pid]; !ok {
				dead = append(dead, pid)
			}
		}

		// Remove dead windows from windowMap
		for _, pid := range dead {
			delete(windowMap, pid)
		}

		// Pack windowList, keeping only alive windows
		newWindowList := make([]win32.Window, 0)
		for _, w := range windowList {
			if _, ok := _windowMap[w.Pid()]; ok {
				newWindowList = append(newWindowList, w)
			}
		}
		newWindowList2 := make([]ClientRow, 0)
		for _, w := range windowList2 {
			if _, ok := _windowMap[w.w.Pid()]; ok {
				newWindowList2 = append(newWindowList2, w)
			}
		}

		// Add new windows to windowMap and windowList
		for pid, row := range _windowMap {
			if _, ok := windowMap[pid]; ok {
				continue
			}
			windowMap[pid] = row
			newWindowList = append(newWindowList, *row.w)
			newWindowList2 = append(newWindowList2, row)
		}

		windowList = newWindowList
		windowList2 = newWindowList2

		w.Invalidate()
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

func BuildClients(theme *material.Theme, btnList []widget.Clickable) (children []layout.FlexChild) {
	for i, w := range windowList {
		c := GameClient{w: &w}
		name, err := c.GetCharacterName()
		if err != nil {
			continue
		}

		_row := ui.ClientRowSyle{
			Name:       name,
			DragHandle: &btnList[i*2],
			Label:      &btnList[i*2+1],
			Theme:      theme,
		}

		children = append(children, layout.Rigid(_row.Layout))

	}

	return
}

func BuildClients2(theme *material.Theme) (children []layout.FlexChild) {
	for _, row := range windowList2 {
		c := GameClient{w: row.w}
		name, err := c.GetCharacterName()
		if err != nil {
			continue
		}

		_row := ui.ClientRowSyle{
			Name:       name,
			DragHandle: row.dragHandle,
			Label:      row.label,
			Theme:      theme,
		}

		children = append(children, layout.Rigid(_row.Layout))

	}

	return
}

func reghk() {
	// Register a desired hotkey.
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyS)
	if err := hk.Register(); err != nil {
		panic("hotkey registration failed")
	}

	// Unregister the hotkey when keydown event is triggered
	for range hk.Keydown() {
		hk.Unregister()
		fmt.Println("Hotkey triggered")
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

func run(window *app.Window) error {
	go reghk()
	window.Option(app.Title("Dofus Manager"))
	window.Option(app.Decorated(false))
	theme := material.NewTheme()
	theme.Palette = defaultPalette

	var ops op.Ops
	var settings widget.Clickable
	var prev widget.Clickable
	var next widget.Clickable
	actions := system.ActionClose | system.ActionMaximize | system.ActionMinimize
	decorations := material.Decorations(theme, &widget.Decorations{}, actions, "Dofus Manager")

	// btnList := make([]widget.Clickable, 10)

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			SetWindowBackground(&gtx, theme)

			window.Perform(decorations.Decorations.Update(gtx))
			decorations.Layout(gtx)

			if prev.Clicked(gtx) {
				fmt.Println("Clicked")
			}

			children := make([]layout.FlexChild, 0)
			// children = BuildClients(theme, btnList)
			children = BuildClients2(theme)
			for _, row := range windowList2 {
				if row.label.Clicked(gtx) {
					fmt.Printf("Bring %s to front\n", row.w.Title())
					go row.w.UIABringToFront()
				}
			}
			// for i := range btnList {
			// 	widx := (i - 1) / 2
			// 	if widx < 0 {
			// 		widx = 0
			// 	}
			// 	if btnList[i].Clicked(gtx) && widx < len(windowList) {
			// 		fmt.Printf("Clicked button %d\n", i)
			// 		fmt.Printf("Bring %s to front\n", windowList[widx].Title())
			// 		go windowList[widx].UIABringToFront()
			// 		// go windowList[widx].BringToFront()
			// 	}
			// }

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

			// Pass the drawing operations to the GPU.
			e.Frame(gtx.Ops)
		}
	}
}
