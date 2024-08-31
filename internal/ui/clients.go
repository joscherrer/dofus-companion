package ui

import (
	"gioui.org/layout"
)

type ClientList struct{}

func (c *ClientList) Layout(gtx layout.Context, rows ...*ClientRowStyle) layout.Dimensions {
	children := make([]layout.FlexChild, len(rows))
	for i, row := range rows {
		children[i] = layout.Rigid(row.Layout)
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx, children...)
}
