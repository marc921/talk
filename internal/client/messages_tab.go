package client

import (
	"github.com/gdamore/tcell/v2"
	"github.com/marc921/talk/internal/types"
)

type MessagesTab struct {
	*BaseComponent
	localUser        *User
	conversation     *types.Conversation
	hasFocus         bool
	mode             Mode
	newMessageBuffer string
}

func NewMessagesTab(base *BaseComponent) *MessagesTab {
	return &MessagesTab{
		BaseComponent: base,
		mode:          ModeNormal,
	}
}

func (c *MessagesTab) CanFocus() bool {
	return true
}

func (c *MessagesTab) Focus(focused bool) {
	if c.hasFocus == focused {
		return
	}
	c.hasFocus = focused
	if focused {
		c.actions <- &ActionSetMode{mode: ModeInsert}
	}
}

func (c *MessagesTab) OnEvent(event any) {
	switch event := event.(type) {
	case *EventSetMode:
		c.mode = event.mode
	case *EventSelectUser:
		c.localUser = event.user
		c.conversation = nil
	case *EventSelectConversation:
		c.conversation = event.conversation
	case *EventFocus:
		c.hasFocus = true
	case *tcell.EventKey:
		if !c.hasFocus {
			return
		}
		switch event.Key() {
		case tcell.KeyEnter:
			if c.mode == ModeInsert {
				c.actions <- &ActionSendMessage{
					localUser:      c.localUser,
					remoteUsername: c.conversation.RemoteUser,
					plaintext:      []byte(c.newMessageBuffer),
				}
				c.newMessageBuffer = ""
				c.actions <- &ActionSetMode{mode: ModeNormal}
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if c.mode == ModeInsert && len(c.newMessageBuffer) > 0 {
				c.newMessageBuffer = c.newMessageBuffer[:len(c.newMessageBuffer)-1]
			}
		case tcell.KeyRune:
			if c.mode == ModeInsert {
				c.newMessageBuffer += string(event.Rune())
			}
		}
	}
}

func (c *MessagesTab) Render() {
	if c.localUser == nil || c.conversation == nil {
		return
	}
	c.drawCursor.Reset()
	style := tcell.StyleDefault.Bold(true).Underline(true)
	if c.hasFocus {
		style = style.Foreground(tcell.ColorDeepSkyBlue)
	}
	c.PrintTextStyle("Messages", style)

	c.drawCursor.Newline()
	for _, message := range c.conversation.Messages {
		if message.From == c.localUser.name {
			c.PrintTextRightAlign(string(message.Plaintext))
		} else {
			c.PrintText(string(message.Plaintext))
		}
		c.drawCursor.Newline()
	}
	if c.hasFocus {
		style = tcell.StyleDefault.Italic(true)
		c.PrintTextStyle(" + New", style)
		if c.mode == ModeInsert {
			c.PrintText(": " + c.newMessageBuffer)
			c.PrintTextStyle("_", tcell.StyleDefault.Blink(true))
		}
		c.drawCursor.Newline()
	}
}
