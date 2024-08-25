package ui

import "gioui.org/layout"

type ClientList struct {
	rows []ClientRowSyle
}

func (c *ClientList) Layout(gtx layout.Context) layout.Dimensions {
	children := make([]layout.FlexChild, len(c.rows))
	for i, row := range c.rows {
		children[i] = layout.Rigid(row.Layout)
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx, children...)
}
