package ui

import (
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"
)

type HotkeyDialog struct {
	Theme *material.Theme
	Text  string
}

func (h *HotkeyDialog) Layout(gtx layout.Context) layout.Dimensions {
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: h.Theme.ContrastBg}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(material.H5(h.Theme, h.Text).Layout),
			layout.Rigid(material.Body1(h.Theme, "or <Esc> to cancel").Layout),
		)
	})
	h.Add(gtx.Ops)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (h *HotkeyDialog) Add(ops *op.Ops) {
	event.Op(ops, h)
}

func (h *HotkeyDialog) Update(gtx layout.Context) (string, bool) {
	for {
		ev, ok := gtx.Event(key.Filter{})
		if !ok {
			return "", false
		}

		if x, ok := ev.(key.Event); ok {
			return string(x.Name), true
		}
	}
}

type HotkeyRow struct {
	Theme      *material.Theme
	ButtonText string
	LabelText  string
	Button     *Clickable
}

func (h *HotkeyRow) Layout(gtx layout.Context) layout.Dimensions {
	hflex := layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}
	tb := Btn(h.Theme, h.Button, h.ButtonText, false)
	label := material.Body1(h.Theme, h.LabelText)
	label.Alignment = text.Alignment(layout.End)
	wLabel := func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Right: 5}.Layout(gtx, label.Layout)
	}
	wButton := func(gtx layout.Context) layout.Dimensions {
		return tb.Layout(gtx)
	}
	return layout.Inset{Bottom: 10}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return hflex.Layout(gtx,
			layout.Flexed(0.5, wLabel),
			layout.Flexed(1, wButton),
		)
	})
}
