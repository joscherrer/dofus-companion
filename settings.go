package main

import (
	"fmt"
	"image"
	"slices"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/joscherrer/dofus-manager/internal/config"
	"github.com/joscherrer/dofus-manager/internal/ui"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SetKeyEvent struct {
	Key string
}

func Title(s string) string {
	caser := cases.Title(language.French)
	return caser.String(s)
}

func runSettings(w *app.Window) error {
	w.Option(app.Title(AppName + " - Settings"))
	w.Option(app.Decorated(false))
	w.Option(app.MaxSize(unit.Dp(400), unit.Dp(400)))
	w.Option(app.MinSize(unit.Dp(400), unit.Dp(400)))
	theme := material.NewTheme()
	theme.Palette = ui.DefaultPalette

	actions := system.ActionClose | system.ActionMaximize | system.ActionMinimize
	decorations := material.Decorations(theme, &widget.Decorations{}, actions, AppName)

	// setKey := false
	setKeyObj := SetKeyEvent{}
	// tbBtn := &ui.Clickable{}
	hkd := &ui.HotkeyDialog{Theme: theme, Text: "Press a key"}
	settingsBtn := make(map[string]*ui.Clickable)

	conf := config.GetConfig()
	changedChan := config.SubscribeToChanged()
	defer config.UnsubscribeFromChanged(changedChan)
	go func() {
		for range changedChan {
			fmt.Println("Refresh settings window")
			conf = config.GetConfig()
			w.Invalidate()
		}
	}()

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			ui.SetWindowBackground(&gtx, theme)

			w.Perform(decorations.Decorations.Update(gtx))
			decoDims := decorations.Layout(gtx)

			children := make([]layout.FlexChild, 0)
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return decoDims
			}))

			if setKeyObj.Key == "" {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := material.H5(theme, "Keybinds")
					title.Alignment = text.Middle
					return layout.UniformInset(10).Layout(gtx, title.Layout)
				}))
				keys := make([]string, 0)
				for k := range conf.Keys {
					keys = append(keys, k)
				}
				slices.Sort(keys)
				for _, k := range keys {
					if _, ok := settingsBtn[k]; !ok {
						settingsBtn[k] = &ui.Clickable{}
					}

					row := ui.HotkeyRow{
						Theme:      theme,
						LabelText:  Title(k) + ": ",
						Button:     settingsBtn[k],
						ButtonText: conf.GetKeyName(k),
					}

					children = append(children, layout.Rigid(row.Layout))
				}
			}

			children = append(children, layout.Flexed(1, layout.Spacer{}.Layout))

			for k := range conf.Keys {
				if settingsBtn[k].Clicked(gtx) {
					setKeyObj = SetKeyEvent{Key: k}
					pauseGlobalHotkeys <- true
				}
			}

			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx, children...)

			if setKeyObj.Key != "" {
				off := op.Offset(image.Pt(0, decoDims.Size.Y)).Push(gtx.Ops)
				gtx.Constraints.Max.Y -= decoDims.Size.Y

				layout.UniformInset(10).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return hkd.Layout(gtx)
				})
				off.Pop()
				if k, ok := hkd.Update(gtx); ok {
					fmt.Println("key: ", k)
					if k != string(key.NameEscape) {
						conf.SetKey(setKeyObj.Key, k)
						config.SetConfig(conf)
					}
					pauseGlobalHotkeys <- false
					setKeyObj = SetKeyEvent{}
					w.Invalidate()
				}

			}

			e.Frame(gtx.Ops)

		}
	}
}
