package ui

import (
	"image/color"

	"gioui.org/widget/material"
)

func argb(c uint32) color.NRGBA {
	return color.NRGBA{A: uint8(c >> 24), R: uint8(c >> 16), G: uint8(c >> 8), B: uint8(c)}
}

func rgb(c uint32) color.NRGBA {
	return argb(0xff000000 | c)
}

var DefaultPalette = material.Palette{
	Fg:         rgb(0xd7dade),
	Bg:         rgb(0x2d2e32),
	ContrastFg: rgb(0xffffff),
	ContrastBg: rgb(0x202224),
}
