package ui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type MenuStyle struct {
	Theme    *material.Theme
	Prev     *widget.Clickable
	Next     *widget.Clickable
	Settings *widget.Clickable
}

func (m *MenuStyle) Layout(gtx layout.Context) layout.Dimensions {
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

	return hFlex.Layout(gtx,
		layout.Rigid(prevBtn),
		layout.Rigid(nextBtn),
		layout.Rigid(settingsBtn),
	)
}
