package client

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/marc921/talk/internal/types"
)

type ConversationsTab struct {
	*BaseComponent
	localUser             *User
	conversations         []*types.Conversation
	hasFocus              bool
	selected              int
	hovered               int
	mode                  Mode
	newConversationBuffer string
}

func NewConversationsTab(base *BaseComponent) *ConversationsTab {
	return &ConversationsTab{
		BaseComponent: base,
		selected:      -1,
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
		c.actions <- &ActionSetMode{mode: ModeNormal}
	}
}

func (c *ConversationsTab) OnEvent(event any) {
	switch event := event.(type) {
	case *EventSetMode:
		c.mode = event.mode
	case *EventSelectUser:
		c.localUser = event.user
		c.conversations = nil
		c.actions <- &ActionFetchMessages{user: c.localUser}
	case *EventAddMessages:
		for _, message := range event.messages {
			remoteUser := ""
			if message.From == c.localUser.name {
				remoteUser = message.To
			} else if message.To == c.localUser.name {
				remoteUser = message.From
			} else {
				c.actions <- &ActionSetError{
					err: fmt.Errorf("received misdirected message: %v", message),
				}
				continue
			}

			foundConversation := false
			for _, conversation := range c.conversations {
				if conversation.RemoteUser == remoteUser {
					conversation.Messages = append(conversation.Messages, message)
					foundConversation = true
					break
				}
			}
			if !foundConversation {
				c.conversations = append(c.conversations, &types.Conversation{
					LocalUser:  c.localUser.name,
					RemoteUser: remoteUser,
					Messages:   []*types.PlainMessage{message},
				})
			}
		}
	case *EventNewConversation:
		c.conversations = append(c.conversations, event.conversation)
		c.actions <- &ActionSelectConversation{conversation: event.conversation}
	case *EventSelectConversation:
		for i, conversation := range c.conversations {
			if conversation.RemoteUser == event.conversation.RemoteUser {
				c.selected = i
				break
			}
		}
		c.actions <- &ActionSwitchTab{tabIndex: TabMessages}
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
			c.hovered = min(c.hovered+1, len(c.conversations))
		case tcell.KeyEnter:
			if c.mode == ModeInsert {
				c.actions <- &ActionCreateConversation{
					localUser:      c.localUser,
					remoteUsername: c.newConversationBuffer,
				}
				c.newConversationBuffer = ""
				c.actions <- &ActionSetMode{mode: ModeNormal}
			} else if c.hovered == len(c.conversations) {
				c.actions <- &ActionSetMode{mode: ModeInsert}
			} else {
				c.actions <- &ActionSelectConversation{conversation: c.conversations[c.hovered]}
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
	for i, conversation := range c.conversations {
		line := fmt.Sprintf(" %d. "+conversation.RemoteUser, i+1)
		style := tcell.StyleDefault
		if i == c.selected {
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
		if c.hovered == len(c.conversations) {
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
