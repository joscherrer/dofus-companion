package ui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type ClientRowSyle struct {
	Name       string
	DragHandle *widget.Clickable
	Label      *widget.Clickable
	Theme      *material.Theme
}

func (c *ClientRowSyle) Layout(gtx layout.Context) layout.Dimensions {
	dragIcon, _ := widget.NewIcon(icons.EditorDragHandle)
	flex := layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceStart,
	}
	label := NewBtn(c.Theme, c.Label, c.Name, true)
	label = AddMargin(label, 0, 10, 10, 10)
	drag := NewIconBtn(c.Theme, c.DragHandle, dragIcon, "move")
	drag = AddMargin(drag, 0, 10, 10, 0)
	return flex.Layout(gtx, layout.Rigid(drag), layout.Flexed(10, label))
}
