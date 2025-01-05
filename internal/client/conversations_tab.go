package client

import (
	"fmt"
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/marc921/talk/internal/types/openapi"
)

type ConversationsTab struct {
	*BaseComponent
	localUser             *User
	hasFocus              bool
	selected              openapi.Username
	hovered               int
	mode                  Mode
	newConversationBuffer string
}

func NewConversationsTab(base *BaseComponent) *ConversationsTab {
	return &ConversationsTab{
		BaseComponent: base,
		selected:      "",
		mode:          ModeNormal,
	}
}

func (c *ConversationsTab) CanFocus() bool {
	return true
}

func (c *ConversationsTab) Focus(focused bool) {
	if c.hasFocus == focused {
		return
	}
	c.hasFocus = focused
	if focused {
		UISingleton.actions <- &ActionSetMode{mode: ModeNormal}
	}
}

func (c *ConversationsTab) OnEvent(event any) {
	switch event := event.(type) {
	case *EventSetMode:
		c.mode = event.mode
	case *EventSelectUser:
		c.localUser = event.user
		UISingleton.actions <- &ActionFetchMessages{user: c.localUser}
	case *EventUpdateUser:
		c.localUser = event.user
	case *EventSelectConversation:
		c.selected = event.conversation.dbConv.RemoteUserName
		UISingleton.actions <- &ActionSwitchTab{tabIndex: TabMessages}
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
			c.hovered = min(c.hovered+1, len(c.localUser.conversations))
		case tcell.KeyEnter:
			if c.mode == ModeInsert {
				UISingleton.actions <- &ActionCreateConversation{
					localUser:      c.localUser,
					remoteUsername: c.newConversationBuffer,
				}
				c.newConversationBuffer = ""
				UISingleton.actions <- &ActionSetMode{mode: ModeNormal}
			} else if c.hovered == len(c.localUser.conversations) {
				UISingleton.actions <- &ActionSetMode{mode: ModeInsert}
			} else {
				remoteUsernames := c.GetSortedRemoteUsernames()
				UISingleton.actions <- &ActionSelectConversation{
					conversation: c.localUser.conversations[remoteUsernames[c.hovered]],
				}
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if c.mode == ModeInsert && len(c.newConversationBuffer) > 0 {
				c.newConversationBuffer = c.newConversationBuffer[:len(c.newConversationBuffer)-1]
			}
		case tcell.KeyRune:
			if c.mode == ModeInsert {
				c.newConversationBuffer += string(event.Rune())
			}
		}
	}
}

func (c *ConversationsTab) GetSortedRemoteUsernames() []openapi.Username {
	keys := make([]string, 0, len(c.localUser.conversations))
	for k := range c.localUser.conversations {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func (c *ConversationsTab) Render() {
	if c.localUser == nil {
		return
	}
	c.drawCursor.Reset()
	style := tcell.StyleDefault.Bold(true).Underline(true)
	if c.hasFocus {
		style = style.Foreground(tcell.ColorDeepSkyBlue)
	}
	c.PrintTextStyle("Conversations", style)

	c.drawCursor.Newline()
	remoteUsernames := c.GetSortedRemoteUsernames()
	for i, remoteUsername := range remoteUsernames {
		line := fmt.Sprintf(" %d. "+remoteUsername, i+1)
		style := tcell.StyleDefault
		if remoteUsername == c.selected {
			style = style.Foreground(tcell.ColorGreen).Bold(true)
		} else if c.hasFocus && i == c.hovered {
			style = style.Foreground(tcell.ColorDeepSkyBlue)
		}
		c.PrintTextStyle(
			line,
			style,
		)
		c.drawCursor.Newline()
	}
	if c.hasFocus {
		style = tcell.StyleDefault.Italic(true)
		if c.hovered == len(c.localUser.conversations) {
			style = style.Foreground(tcell.ColorDeepSkyBlue)
		}
		c.PrintTextStyle(" + New", style)
		if c.mode == ModeInsert {
			c.PrintText(": " + c.newConversationBuffer)
			c.PrintTextStyle("_", tcell.StyleDefault.Blink(true))
		}
		c.drawCursor.Newline()
	}
}
