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
	if UISingleton == nil {
		panic("UISingleton is nil")
	}
	UISingleton.actions <- new(ActionListUsers)
	return c
}

func (c *UsersTab) CanFocus() bool {
	return true
}

func (c *UsersTab) Focus(focused bool) {
	if c.hasFocus == focused {
		return
	}
	c.hasFocus = focused
	if focused {
		UISingleton.actions <- &ActionSetMode{mode: ModeNormal}
	}
}

func (c *UsersTab) OnEvent(event any) {
	switch event := event.(type) {
	case *EventSetMode:
		c.mode = event.mode
	case *EventSetUsers:
		c.users = event.users
	case *EventNewUser:
		c.users = append(c.users, event.user)
		UISingleton.actions <- &ActionSelectUser{user: event.user}
	case *EventSelectUser:
		for i, user := range c.users {
			if user.name == event.user.name {
				c.selected = i
				break
			}
		}
		UISingleton.actions <- &ActionSwitchTab{tabIndex: TabConversations}
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
				UISingleton.actions <- &ActionCreateUser{username: c.newUsernameBuffer}
				c.newUsernameBuffer = ""
				UISingleton.actions <- &ActionSetMode{mode: ModeNormal}
			} else if c.hovered == len(c.users) {
				UISingleton.actions <- &ActionSetMode{mode: ModeInsert}
			} else {
				UISingleton.actions <- &ActionSelectUser{user: c.users[c.hovered]}
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
	c.drawCursor.Reset()
	style := tcell.StyleDefault.Bold(true).Underline(true)
	if c.hasFocus {
		style = style.Foreground(tcell.ColorDeepSkyBlue)
	}
	c.PrintTextStyle("Users", style)

	c.drawCursor.Newline()
	for i, user := range c.users {
		line := fmt.Sprintf(" %d. %s", i+1, user.name)
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
