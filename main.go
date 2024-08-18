package main

import (
	"image"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/joscherrer/dofus-manager/assets"
	"github.com/joscherrer/dofus-manager/internal/ui"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"gioui.org/widget"
	"gioui.org/widget/material"
)

var windowList []Window

func WatchWindows(w *app.Window) {
	for {
		windowList, _ = GetDofusWindows()
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
		name, err := c.w.GetWindowText()
		if err != nil {
			continue
		}

		// btn := ui.NewButton(theme, &btnList[i], name)
		btn := ui.NewBtn(theme, &btnList[i], name)
		btn = ui.AddMargin(btn, 0, 10, 10, 10)

		children = append(children, layout.Rigid(btn))
	}
	return
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
	window.Option(app.Title("Dofus Manager"))
	theme := material.NewTheme()
	theme.Palette = defaultPalette

	var ops op.Ops
	var settings widget.Clickable
	var prev widget.Clickable
	var next widget.Clickable
	btnList := make([]widget.Clickable, 10)

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)

			SetWindowBackground(&gtx, theme)

			children := make([]layout.FlexChild, 0)

			if windowList != nil {
				children = BuildClients(theme, btnList)
			}

			settingsBtn := ui.NewBtn(theme, &settings, "Settings")
			settingsBtn = ui.AddMargin(settingsBtn, 0, 10, 10, 10)

			prevIcon, _ := widget.NewIcon(icons.NavigationArrowBack)
			prevBtn := ui.NewIconBtn(theme, &prev, prevIcon, "prev")
			prevBtn = ui.AddMargin(prevBtn, 0, 10, 0, 10)

			nextIcon, _ := widget.NewIcon(icons.NavigationArrowForward)
			nextBtn := ui.NewIconBtn(theme, &next, nextIcon, "next")
			nextBtn = ui.AddMargin(nextBtn, 0, 10, 0, 10)

			testBtn := ui.NewAssetBtn(theme, &next, &assets.Image_arrow_up, "up")
			testBtn = ui.AddMargin(testBtn, 0, 10, 0, 10)

			hFlex := layout.Flex{
				Axis:    layout.Horizontal,
				Spacing: layout.SpaceStart,
			}

			tmenu := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return hFlex.Layout(gtx,
					layout.Flexed(1, settingsBtn),
					layout.Rigid(prevBtn),
					layout.Rigid(nextBtn),
					layout.Rigid(testBtn),
				)
			})
			children = append(children, tmenu)

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
