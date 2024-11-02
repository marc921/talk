package client

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type SeparatorLine struct {
	*BaseComponent
}

func NewSeparatorLine(base *BaseComponent) *SeparatorLine {
	c := &SeparatorLine{
		BaseComponent: base,
	}
	return c
}

func (c *SeparatorLine) CanFocus() bool {
	return false
}

func (c *SeparatorLine) Focus(focused bool) {
	// A separator line does not need to be focused
}

func (c *SeparatorLine) OnEvent(event any) {
	// A separator line does not need to handle events
}

func (c *SeparatorLine) Render() {
	c.drawCursor.Reset(c.rect)
	var r rune
	if c.rect.Width == 1 {
		r = '│' // '│', U+2502, BOX DRAWINGS LIGHT VERTICAL
	} else if c.rect.Height == 1 {
		r = '─' // '─', U+2500, BOX DRAWINGS LIGHT HORIZONTAL
	} else {
		c.actions <- &ActionSetError{
			err: fmt.Errorf("invalid separator line dimensions: %v", c.rect),
		}
	}
	for y := c.rect.Y; y < c.rect.Y+c.rect.Height; y++ {
		for x := c.rect.X; x < c.rect.X+c.rect.Width; x++ {
			c.screen.SetContent(x, y, r, nil, tcell.StyleDefault)
		}
	}
}
