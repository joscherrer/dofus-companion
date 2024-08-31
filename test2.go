package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/joscherrer/dofus-manager/internal/ui"
)

func run2(window *app.Window) error {
	var ops op.Ops

	const mime = "MyMime"
	var drop int
	drag := &ui.Draggable{Type: mime}
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			// mime is the type used to match drag and drop operations.
			// It could be left empty in this example.
			// widget lays out the drag and drop handlers and processes
			// the transfer events.
			w := func(gtx layout.Context) layout.Dimensions {
				sz := image.Pt(10, 10) // drag area
				r := clip.Rect{Max: sz}.Push(gtx.Ops)
				paint.ColorOp{Color: color.NRGBA{R: 0x80, A: 0xFF}}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				r.Pop()
				return layout.Dimensions{Size: sz}
			}
			drag.Layout(gtx, w, w)
			// drag must respond with an Offer event when requested.
			// Use the drag method for this.
			if m, ok := drag.Update(gtx); ok {
				drag.Offer(gtx, m, io.NopCloser(strings.NewReader("hello world")))
			}

			// Setup the area for drops.

			ds := clip.Rect{
				Min: image.Pt(20, 20),
				Max: image.Pt(40, 40),
			}.Push(gtx.Ops)
			event.Op(gtx.Ops, &drop)
			paint.ColorOp{Color: color.NRGBA{A: 0x80, R: 0xFF}}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			ds.Pop()

			// Check for the received data.
			for {
				ev, ok := gtx.Event(transfer.TargetFilter{Target: &drop, Type: mime})
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
			e.Frame(gtx.Ops)

		}
	}
}

func draw(gtx layout.Context) {
}
