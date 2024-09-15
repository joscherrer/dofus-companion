package ui

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type ClientRowStyle struct {
	Name       string
	Label      *Clickable
	DragHandle *Clickable
	Draggable  *Draggable
	Theme      *material.Theme
	index      int
	Released   bool
}

func (c *ClientRowStyle) Layout(gtx layout.Context) layout.Dimensions {
	c.Released = false
	dragIcon, _ := widget.NewIcon(icons.EditorDragHandle)
	flex := layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceStart,
	}
	label := NewBtn(c.Theme, c.Label, c.Name, false)
	label = AddMargin(label, 0, 10, 0, 10)

	drag := func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(8).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return dragIcon.Layout(gtx, c.Theme.Fg)
		})
	}

	r := func(gtx layout.Context) layout.Dimensions {
		return flex.Layout(gtx, layout.Rigid(drag), layout.Flexed(10, label))
	}

	draggable := func(gtx layout.Context) layout.Dimensions {
		return c.Draggable.Layout(gtx, drag, r)
	}
	dims := flex.Layout(gtx, layout.Rigid(draggable), layout.Flexed(10, label))
	if _, ok := c.Draggable.Update(gtx); ok {
		gtx.Execute(op.InvalidateCmd{})
	}

	return dims
}
