package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

func AddMargin(w layout.Widget, t, b, l, r int) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(t),
			Bottom: unit.Dp(b),
			Left:   unit.Dp(l),
			Right:  unit.Dp(r),
		}.Layout(gtx, w)
	}
}
