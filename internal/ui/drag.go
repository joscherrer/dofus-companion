// package ui
//
// import (
// 	"image/color"
//
// 	"gioui.org/io/semantic"
// 	"gioui.org/layout"
// 	"gioui.org/op/clip"
// 	"gioui.org/op/paint"
// )
//
// type DragHandle struct {
// 	Button     *Clickable
// 	Background color.NRGBA
// }
//
// func (d DragHandle) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
// 	min := gtx.Constraints.Min
// 	return d.Button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
// 		semantic.Button.Add(gtx.Ops)
// 		return layout.Background{}.Layout(gtx,
// 			func(gtx layout.Context) layout.Dimensions {
// 				defer clip.Rect{Max: min}.Push(gtx.Ops).Pop()
// 				paint.Fill(gtx.Ops, d.Background)
// 				return layout.Dimensions{Size: gtx.Constraints.Min}
// 			},
// 			func(gtx layout.Context) layout.Dimensions {
// 				return layout.Dimensions{}
// 			},
// 		)
// 	})
// }

package ui

import (
	"fmt"
	"io"
	"reflect"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/transfer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// Draggable makes a widget draggable.
type Draggable struct {
	// Type contains the MIME type and matches transfer.SourceOp.
	Type string

	drag  gesture.Drag
	click f32.Point
	pos   f32.Point
}

func (d *Draggable) Layout(gtx layout.Context, w, drag layout.Widget) layout.Dimensions {
	if !gtx.Enabled() {
		return w(gtx)
	}
	dims := w(gtx)

	stack := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	d.drag.Add(gtx.Ops)
	event.Op(gtx.Ops, d)
	stack.Pop()

	if drag != nil && d.drag.Pressed() {
		rec := op.Record(gtx.Ops)
		op.Offset(d.pos.Round()).Add(gtx.Ops)
		drag(gtx)
		op.Defer(gtx.Ops, rec.Stop())
	}

	return dims
}

// Dragging returns whether d is being dragged.
func (d *Draggable) Dragging() bool {
	return d.drag.Dragging()
}

// Update the draggable and returns the MIME type for which the Draggable was
// requested to offer data, if any
func (d *Draggable) Update(gtx layout.Context) (mime string, requested bool) {
	pos := d.pos
	for {
		ev, ok := d.drag.Update(gtx.Metric, gtx.Source, gesture.Both)
		if !ok {
			break
		}
		switch ev.Kind {
		case pointer.Press:
			d.click = ev.Position
			pos = f32.Point{}
		case pointer.Drag, pointer.Release:
			pos = ev.Position.Sub(d.click)
		}
	}
	d.pos = pos

	for {
		e, ok := gtx.Event(transfer.SourceFilter{Target: d, Type: d.Type})
		if !ok {
			break
		}
		fmt.Println(reflect.TypeOf(e).String())
		if e, ok := e.(transfer.RequestEvent); ok {
			return e.Type, true
		}
		fmt.Println("e: ", e)
		fmt.Println("d: ", d)
	}
	return "", false
}

// Offer the data ready for a drop. Must be called after being Requested.
// The mime must be one in the requested list.
func (d *Draggable) Offer(gtx layout.Context, mime string, data io.ReadCloser) {
	gtx.Execute(transfer.OfferCmd{Tag: d, Type: mime, Data: data})
}

// Pos returns the drag position relative to its initial click position.
func (d *Draggable) Pos() f32.Point {
	return d.pos
}
