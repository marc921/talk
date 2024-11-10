package client

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

type EventUpdateUser struct {
	user *User
}

type EventSetError struct {
	err error
}

type EventSelectConversation struct {
	conversation *Conversation
}

type EventSwitchTab struct {
	tabIndex TabIndex
}
