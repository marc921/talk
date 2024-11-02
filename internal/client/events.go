package client

import "github.com/marc921/talk/internal/types"

type EventSetMode struct {
	mode Mode
}

type EventSetUsers struct {
	users []*User
}

type EventSelectUser struct {
	user *User
}

type EventFocus struct {
}

type EventNewUser struct {
	user *User
}

type EventAddMessages struct {
	messages []*types.PlainMessage
}

type EventSetError struct {
	err error
}

type EventNewConversation struct {
	conversation *types.Conversation
}

type EventSelectConversation struct {
	conversation *types.Conversation
}
