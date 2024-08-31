package main

import (
	"strings"

	"github.com/joscherrer/dofus-manager/internal/window"
)

type GameClient struct {
	w *window.Window
}

func (c *GameClient) GetCharacterName() (s string, err error) {
	winTitle, _ := c.w.GetWindowText()
	s, _, _ = strings.Cut(winTitle, " - ")
	return
}
