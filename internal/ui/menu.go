package ui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/joscherrer/dofus-manager/internal/window"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type MenuStyle struct {
	Theme    *material.Theme
	Prev     *Clickable
	Next     *Clickable
	Settings *Clickable
	Save     *Clickable
}

func (m *MenuStyle) Layout(gtx layout.Context) layout.Dimensions {
	// saveIcon, _ := widget.NewIcon(icons.ContentSave)
	// saveBtn := NewIconBtn(m.Theme, m.Save, saveIcon, "Save")
	// saveBtn = AddMargin(saveBtn, 0, 10, 5, 10)

	settingsIcon, _ := widget.NewIcon(icons.ActionSettings)
	settingsBtn := NewIconBtn(m.Theme, m.Settings, settingsIcon, "Settings")
	settingsBtn = AddMargin(settingsBtn, 0, 10, 5, 10)

	prevIcon, _ := widget.NewIcon(icons.NavigationArrowBack)
	prevBtn := NewIconBtn(m.Theme, m.Prev, prevIcon, "prev")
	prevBtn = AddMargin(prevBtn, 0, 10, 10, 5)

	nextIcon, _ := widget.NewIcon(icons.NavigationArrowForward)
	nextBtn := NewIconBtn(m.Theme, m.Next, nextIcon, "next")
	nextBtn = AddMargin(nextBtn, 0, 10, 0, 0)

	hFlex := layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceStart,
	}

	if m.Prev.Clicked(gtx) {
		window.GetManager().FocusPrev()
	}

	if m.Next.Clicked(gtx) {
		window.GetManager().FocusNext()
	}

	// if m.Save.Clicked(gtx) {
	// 	window.GetManager().SaveOrder()
	// }

	return hFlex.Layout(gtx,
		layout.Rigid(prevBtn),
		layout.Rigid(nextBtn),
		// layout.Rigid(saveBtn),
		layout.Rigid(settingsBtn),
	)
}
