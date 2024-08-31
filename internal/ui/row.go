package ui

import (
	"fmt"
	"io"
	"strings"

	"gioui.org/io/transfer"
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
}

func (c *ClientRowStyle) Layout(gtx layout.Context) layout.Dimensions {
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
	if m, ok := c.Draggable.Update(gtx); ok {
		fmt.Println("Offering data")
		c.Draggable.Offer(gtx, m, io.NopCloser(strings.NewReader("test")))
		gtx.Execute(op.InvalidateCmd{})
	}

	// event.Op(gtx.Ops, &c)

	for {
		ev, ok := gtx.Event(transfer.TargetFilter{Target: &c, Type: "ClientRow"})
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

	return dims
}
