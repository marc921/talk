package client

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Header struct {
	*BaseComponent
	mode            Mode
	currentUsername *string
}

func NewHeader(base *BaseComponent) *Header {
	c := &Header{
		BaseComponent: base,
		mode:          ModeNormal,
	}
	return c
}

func (c *Header) CanFocus() bool {
	return false
}

func (c *Header) Focus(focused bool) {
	// The header does not need to be focused
}

func (c *Header) OnEvent(event any) {
	switch event := event.(type) {
	case *EventSetMode:
		c.mode = event.mode
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyEscape:
			UISingleton.actions <- &ActionSetMode{mode: ModeNormal}
		case tcell.KeyRune:
			if c.mode == ModeInsert {
				// Insert mode does not affect the header
				return
			}
			switch event.Rune() {
			case 'q':
				UISingleton.actions <- new(ActionQuit)
			}
		}
	}
}

func (c *Header) Render() {
	c.drawCursor.Reset()
	width, _ := c.screen.Size()

	// Print the left part of the header
	c.PrintTextStyle(" Talk ", tcell.StyleDefault.Bold(true).Italic(true).Foreground(tcell.ColorGold))
	headerParts := []string{
		"│ Mode: " + string(c.mode),
	}
	if c.currentUsername != nil {
		headerParts = append(headerParts, "User: "+*c.currentUsername)
	}
	c.PrintText(strings.Join(headerParts, " │ "))

	// Print the right part of the header
	c.PrintTextRightAlign("│ [Q]uit ")

	// Print the line separator
	c.drawCursor.Newline()
	c.PrintText(strings.Repeat("-", width))
	c.drawCursor.Newline()
}
