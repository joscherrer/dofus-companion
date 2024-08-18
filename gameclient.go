package main

import (
	"fmt"
	"strings"
)

type GameClient struct {
	w *Window
}

func (c *GameClient) GetCharacterName() (s string, err error) {
	winTitle, _ := c.w.GetWindowText()
	before, after, found := strings.Cut(winTitle, " - ")
	if found && strings.Contains(after, "Dofus") {
		s = before
	} else {
		err = fmt.Errorf("Could not find character name in window title")
	}
	return
}
