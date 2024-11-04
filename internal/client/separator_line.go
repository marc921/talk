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
	c.drawCursor.Reset()
	var r rune
	if c.bounds.Width == 1 {
		r = '│' // '│', U+2502, BOX DRAWINGS LIGHT VERTICAL
	} else if c.bounds.Height == 1 {
		r = '─' // '─', U+2500, BOX DRAWINGS LIGHT HORIZONTAL
	} else {
		c.actions <- &ActionSetError{
			err: fmt.Errorf("invalid separator line dimensions: %v", c.bounds),
		}
	}
	for y := c.bounds.Top; y < c.bounds.Top+c.bounds.Height; y++ {
		for x := c.bounds.Left; x < c.bounds.Left+c.bounds.Width; x++ {
			c.screen.SetContent(x, y, r, nil, tcell.StyleDefault)
		}
	}
}
