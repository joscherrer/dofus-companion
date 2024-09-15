package ui

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/joscherrer/dofus-manager/internal/window"
	"golang.org/x/sys/windows"
)

type ClientList struct {
	MovedRowIndex  int
	MovedRowOffset int
	draggingPos    f32.Point
	draggingIndex  int
	Theme          *material.Theme
	Draggables     map[windows.Handle]*Draggable
	Clickables     map[windows.Handle]*Clickable
}

func (c *ClientList) Layout(gtx layout.Context) layout.Dimensions {
	manager := window.GetManager()
	hwnds := manager.GetSortedHwnds()
	children := make([]layout.FlexChild, len(hwnds))
	rows := make([]*ClientRowStyle, len(hwnds))

	for i, hwnd := range hwnds {
		if _, ok := c.Draggables[hwnd]; !ok {
			c.Draggables[hwnd] = &Draggable{}
		}
		if _, ok := c.Clickables[hwnd]; !ok {
			c.Clickables[hwnd] = &Clickable{}
		}
		labelTxt := manager.GetWindowLabel(hwnd)
		rows[i] = &ClientRowStyle{
			Theme:     c.Theme,
			Draggable: c.Draggables[hwnd],
			Label:     c.Clickables[hwnd],
			Name:      labelTxt,
		}
		children[i] = layout.Rigid(rows[i].Layout)
		if c.Clickables[hwnd].Clicked(gtx) {
			manager.BringToFront(hwnd)
		}
		c.Clickables[hwnd].Active = manager.ActiveHwnd() == hwnd
	}

	dims := layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx, children...)

	rowHeight := .0

	if len(rows) != 0 {
		rowHeight = float64(dims.Size.Y / len(rows))
	}
	if c.draggingPos.Y != 0 {
		initialPos := c.draggingIndex * int(rowHeight)
		moveStep := int(math.Floor((float64(c.draggingPos.Y) + rowHeight/2) / rowHeight))
		if c.draggingIndex+moveStep < 0 {
			moveStep = -c.draggingIndex
		} else if c.draggingIndex+moveStep > len(rows) {
			moveStep = len(rows) - c.draggingIndex
		}
		offset := op.Offset(image.Pt(0, initialPos+(moveStep*int(rowHeight)))).Push(gtx.Ops)
		sep := clip.Rect{Max: image.Pt(gtx.Constraints.Max.X, 2)}.Push(gtx.Ops)
		paint.ColorOp{Color: c.Theme.Fg}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		sep.Pop()
		offset.Pop()
	}

	c.MovedRowIndex = -1
	c.MovedRowOffset = 0
	c.draggingIndex = -1
	c.draggingPos = f32.Point{}

	for i, row := range rows {
		y := float64(row.Draggable.Pos().Y)
		if !row.Draggable.Dragging() && y != 0 {
			row.Draggable.ResetPos()
			c.MovedRowIndex = i
			c.MovedRowOffset = int(math.Floor((y + rowHeight/2) / rowHeight))
			if y > 0 && c.MovedRowOffset > 0 {
				c.MovedRowOffset--
			}
			manager.MoveWindow(hwnds[i], c.MovedRowOffset)
			gtx.Execute(op.InvalidateCmd{})
		}
		if row.Draggable.Dragging() {
			c.draggingIndex = i
			c.draggingPos = row.Draggable.Pos()
		}
	}

	return dims
}
