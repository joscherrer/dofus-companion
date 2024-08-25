package main

import (
	"strings"

	"github.com/joscherrer/dofus-manager/internal/win32"
)

type GameClient struct {
	w *win32.Window
}

func (c *GameClient) GetCharacterName() (s string, err error) {
	winTitle, _ := c.w.GetWindowText()
	s, _, _ = strings.Cut(winTitle, " - ")
	return
}
