package main

import (
	"flag"
	"fmt"
	"log"
	"maps"
	"os"
	"runtime/pprof"
	"sync"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/joscherrer/dofus-manager/internal/config"
	"github.com/joscherrer/dofus-manager/internal/ui"
	"github.com/joscherrer/dofus-manager/internal/window"
	"golang.org/x/sys/windows"

	"gioui.org/widget"
	"gioui.org/widget/material"

	"golang.design/x/hotkey"
)

var (
	clientList = &ui.ClientList{
		Draggables: make(map[windows.Handle]*ui.Draggable),
		Clickables: make(map[windows.Handle]*ui.Clickable),
	}
	windowListMutex = sync.Mutex{}
	focusIdx        = 0
	menu            = &ui.MenuStyle{
		Prev:     &ui.Clickable{},
		Next:     &ui.Clickable{},
		Settings: &ui.Clickable{},
		Save:     &ui.Clickable{},
	}
	setEditor = ui.SetEditor{
		Editor: &widget.Editor{
			Filter: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMOPQRSTUVWXYZ0123456789_",
			Submit: true,
		},
		Save: &ui.Clickable{},
	}
	pauseGlobalHotkeys = make(chan bool)
	debug              = flag.Bool("debug", false, "enable debug mode")
)

const AppName = "Dofus Companion"

func WatchWindows(w *app.Window) {
	manager := window.GetManager()
	for {
		if !manager.UpdateWindows() {
			continue
		}
		w.Invalidate()
	}
}

func registerHk(key hotkey.Key) (hk *hotkey.Hotkey) {
	hk = hotkey.New([]hotkey.Modifier{}, key)
	if err := hk.Register(); err != nil {
		log.Printf("hotkey registration failed: %v\n", err)
	}
	return
}

func invlidateOnConfigChange(w *app.Window) {
	ch := config.SubscribeToChanged()
	defer config.UnsubscribeFromChanged(ch)
	for range ch {
		if !window.GetManager().SetKeyUpdated {
			continue
		}
		w.Invalidate()
	}
}

func hkLoop(w *app.Window) {
	c := config.GetConfig()
	ch := config.SubscribeToChanged()
	defer config.UnsubscribeFromChanged(ch)
	paused := false

	prev := registerHk(c.GetKey("previous"))
	next := registerHk(c.GetKey("next"))

	for {
		select {
		case <-prev.Keydown():
			window.GetManager().FocusPrev()
			w.Invalidate()
		case <-next.Keydown():
			window.GetManager().FocusNext()
			w.Invalidate()
		case <-ch:
			newConfig := config.GetConfig()
			if maps.Equal(c.Keys, newConfig.Keys) {
				continue
			}
			fmt.Println("Config changed, re-registerings hotkeys")
			prev.Unregister()
			next.Unregister()

			c = newConfig

			prev = registerHk(c.GetKey("previous"))
			next = registerHk(c.GetKey("next"))
		case p := <-pauseGlobalHotkeys:
			if p && !paused {
				prev.Unregister()
				next.Unregister()
				paused = true
			}
			if !p && paused {
				prev = registerHk(c.GetKey("previous"))
				next = registerHk(c.GetKey("next"))
				paused = false
			}
		}
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

func run(w *app.Window) error {
	go hkLoop(w)
	go invlidateOnConfigChange(w)

	w.Option(app.Title(AppName))
	w.Option(app.Decorated(false))
	theme := material.NewTheme()
	theme.Palette = ui.DefaultPalette
	menu.Theme = theme
	setEditor.Theme = theme
	clientList.Theme = theme
	wSettingsOpen := false
	var wSettings *app.Window

	var ops op.Ops
	actions := system.ActionClose | system.ActionMaximize | system.ActionMinimize
	decorations := material.Decorations(theme, &widget.Decorations{}, actions, AppName)
	if *debug {
		cpuprofile := "cpu.prof"
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			ui.SetWindowBackground(&gtx, theme)

			w.Perform(decorations.Decorations.Update(gtx))

			children := make([]layout.FlexChild, 0)
			children = append(children, layout.Rigid(decorations.Layout))
			children = append(children, layout.Rigid(setEditor.Layout))
			children = append(children, layout.Flexed(1, layout.Spacer{}.Layout))
			children = append(children, layout.Rigid(clientList.Layout))
			children = append(children, layout.Rigid(menu.Layout))

			flex := layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}

			manager := window.GetManager()
			if setEditor.Save.Clicked(gtx) {
				manager.SaveOrder(setEditor.Editor.Text())
			}

			if ev, ok := setEditor.Editor.Update(gtx); ok {
				switch ev.(type) {
				case widget.SubmitEvent:
					manager.SaveOrder(setEditor.Editor.Text())
				}
			}

			if manager.SetKeyUpdated {
				setEditor.Editor.SetText(manager.SetKey)
				manager.SetKeyUpdated = false
			}

			if menu.Settings.Clicked(gtx) {
				if wSettingsOpen {
					wSettings.Perform(system.ActionRaise)
				} else {
					wSettings = new(app.Window)
					wSettingsOpen = true
					go func() {
						err := runSettings(wSettings)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println("Settings window closed")
						wSettingsOpen = false
					}()
				}
			}

			flex.Layout(gtx, children...)

			e.Frame(gtx.Ops)
		}
	}
}
