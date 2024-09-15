package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
)

func SetWindowBackground(gtx *layout.Context, theme *material.Theme) {
	macro := op.Record(gtx.Ops)
	rect := image.Rectangle{
		Max: image.Point{
			X: gtx.Constraints.Max.X,
			Y: gtx.Constraints.Max.Y,
		},
	}
	paint.FillShape(gtx.Ops, theme.Bg, clip.Rect(rect).Op())
	// paint.FillShape(gtx.Ops, rgb(0x2d2e32), clip.Rect(rect).Op())
	background := macro.Stop()

	background.Add(gtx.Ops)
}
