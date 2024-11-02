package client

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type UsersTab struct {
	*BaseComponent
	users             []*User
	hasFocus          bool
	selected          int
	hovered           int
	mode              Mode
	newUsernameBuffer string
}

func NewUsersTab(base *BaseComponent) *UsersTab {
	c := &UsersTab{
		BaseComponent: base,
		selected:      -1,
		mode:          ModeNormal,
	}
	c.actions <- new(ActionListUsers)
	return c
}

func (c *UsersTab) CanFocus() bool {
	return true
}
func (c *UsersTab) Focus(focused bool) {
	c.hasFocus = focused
}

func (c *UsersTab) OnEvent(event any) {
	switch event := event.(type) {
	case *EventSetMode:
		c.mode = event.mode
	case *EventSetUsers:
		c.users = event.users
	case *EventNewUser:
		c.users = append(c.users, event.user)
		c.actions <- &ActionSelectUser{user: event.user}
	case *EventSelectUser:
		for i, user := range c.users {
			if user.name == event.user.name {
				c.selected = i
				break
			}
		}
	case *EventFocus:
		c.hasFocus = true
	case *tcell.EventKey:
		if !c.hasFocus {
			return
		}
		switch event.Key() {
		case tcell.KeyUp:
			c.hovered = max(c.hovered-1, 0)
		case tcell.KeyDown:
			c.hovered = min(c.hovered+1, len(c.users))
		case tcell.KeyEnter:
			if c.mode == ModeInsert {
				c.actions <- &ActionCreateUser{username: c.newUsernameBuffer}
				c.newUsernameBuffer = ""
				c.actions <- &ActionSetMode{mode: ModeNormal}
			} else if c.hovered == len(c.users) {
				c.actions <- &ActionSetMode{mode: ModeInsert}
			} else {
				c.actions <- &ActionSelectUser{user: c.users[c.hovered]}
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if c.mode == ModeInsert && len(c.newUsernameBuffer) > 0 {
				c.newUsernameBuffer = c.newUsernameBuffer[:len(c.newUsernameBuffer)-1]
			}
		case tcell.KeyRune:
			if c.mode == ModeInsert {
				c.newUsernameBuffer += string(event.Rune())
			}
		}
	}
}

func (c *UsersTab) Render() {
	c.drawCursor.Reset(c.rect)
	style := tcell.StyleDefault.Bold(true).Underline(true)
	if c.hasFocus {
		style = style.Foreground(tcell.ColorDeepSkyBlue)
	}
	c.PrintTextStyle("Users", style)

	c.drawCursor.Newline()
	for i, user := range c.users {
		line := fmt.Sprintf(" %d. "+user.name, i+1)
		style := tcell.StyleDefault
		if i == c.selected {
			style = style.Foreground(tcell.ColorGreen).Bold(true)
		} else if c.hasFocus && i == c.hovered {
			style = style.Foreground(tcell.ColorDeepSkyBlue)
		}
		c.PrintTextStyle(line, style)
		c.drawCursor.Newline()
	}
	if c.hasFocus {
		style = tcell.StyleDefault.Italic(true)
		if c.hovered == len(c.users) {
			style = style.Foreground(tcell.ColorDeepSkyBlue)
		}
		c.PrintTextStyle(" + New", style)
		if c.mode == ModeInsert {
			c.PrintText(": " + c.newUsernameBuffer)
			c.PrintTextStyle("_", tcell.StyleDefault.Blink(true))
		}
		c.drawCursor.Newline()
	}
}
