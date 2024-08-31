package ui

import (
	"gioui.org/widget"
)

type Clickable struct {
	widget.Clickable
	Active bool
}
