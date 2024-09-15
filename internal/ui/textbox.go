package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/joscherrer/dofus-manager/internal/f32color"
)

type TextBoxStyle struct {
	Theme  *material.Theme
	Editor *widget.Editor
	Hint   string
}

func (e TextBoxStyle) Layout(gtx layout.Context) layout.Dimensions {
	editorStyle := material.Editor(e.Theme, e.Editor, e.Hint)
	editorStyle.HintColor = f32color.MulAlpha(e.Theme.Fg, 0x55)
	editorStyle.SelectionColor = f32color.Hovered(e.Theme.Bg)
	dims := image.Point{gtx.Constraints.Max.X, 38}
	rect := clip.UniformRRect(image.Rectangle{Max: dims}, 5).Push(gtx.Ops)
	paint.ColorOp{Color: e.Theme.ContrastBg}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	layout.UniformInset(8).Layout(gtx, editorStyle.Layout)
	rect.Pop()
	return layout.Dimensions{Size: dims}
}
