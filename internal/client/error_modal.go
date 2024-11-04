package client

import (
	"github.com/gdamore/tcell/v2"
)

var ErrStyle = tcell.StyleDefault.Bold(true).Foreground(tcell.ColorRed)

type ErrorModal struct {
	*BaseComponent
	err error
}

func NewErrorModal(base *BaseComponent) *ErrorModal {
	return &ErrorModal{
		BaseComponent: base,
	}
}

func (c *ErrorModal) CanFocus() bool {
	return false
}

func (c *ErrorModal) Focus(focused bool) {
	// The error modal does not respond to external focus events
}

func (c *ErrorModal) OnEvent(event any) {
	switch event := event.(type) {
	case *EventSetError:
		c.err = event.err
	case *tcell.EventKey:
		if c.err == nil {
			return
		}
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyEscape:
			c.actions <- &ActionSetError{err: nil}
		}
	}
}

func (c *ErrorModal) Render() {
	if c.err == nil {
		return
	}
	width, height := c.screen.Size()

	errHeight := len(c.err.Error())/(width-6) + 1
	boxBounds := &Rect{
		Left:   2,
		Top:    height/2 - errHeight/2 - 1,
		Width:  width - 4,
		Height: errHeight + 2,
	}

	c.PrintBox(boxBounds, ErrStyle)
	c.drawCursor.MoveTo(boxBounds.Left, boxBounds.Top)
	c.PrintTextCentered(" ERROR ", ErrStyle)
	c.drawCursor.MoveTo(boxBounds.Left, boxBounds.Top+boxBounds.Height-1)
	c.PrintTextCentered(" OK [Enter] ", ErrStyle)
	c.SetBounds(boxBounds.Shrink(1))
	c.drawCursor.Reset()
	errMessage := c.err.Error()
	for len(errMessage) > 0 {
		if len(errMessage) > width-6 {
			c.PrintTextStyle(errMessage[:width-6], ErrStyle)
			errMessage = errMessage[width-6:]
		} else {
			c.PrintTextStyle(errMessage, ErrStyle)
			errMessage = ""
		}
		c.drawCursor.Newline()
	}
}
