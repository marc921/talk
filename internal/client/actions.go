package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/marc921/talk/internal/types"
)

type Action interface {
	Do(ctx context.Context, u *UI) error
	String() string
}

var ErrQuit = errors.New("quit")

type ActionQuit struct{}

func (a *ActionQuit) Do(ctx context.Context, u *UI) error {
	u.Quit()
	return ErrQuit
}

func (a *ActionQuit) String() string {
	return "Quit"
}

type ActionSetError struct {
	err error
}

func (a *ActionSetError) Do(ctx context.Context, u *UI) error {
	u.drawer.OnEvent(&EventSetError{err: a.err})
	return nil
}

func (a *ActionSetError) String() string {
	return "SetError"
}

type ActionSetMode struct {
	mode Mode
}

func (a *ActionSetMode) Do(ctx context.Context, u *UI) error {
	u.drawer.OnEvent(&EventSetMode{mode: a.mode})
	return nil
}

func (a *ActionSetMode) String() string {
	return "Mode"
}

type ActionListUsers struct{}

func (a *ActionListUsers) Do(ctx context.Context, u *UI) error {
	usernames, err := u.storage.ListUsers()
	if err != nil {
		return fmt.Errorf("storage.ListUsers: %w", err)
	}
	users := make([]*User, 0, len(usernames))
	for _, username := range usernames {
		user, err := NewUser(username, u.storage, u.openapiClient)
		if err != nil {
			return fmt.Errorf("NewUser: %w", err)
		}
		users = append(users, user)
	}
	u.drawer.OnEvent(&EventSetUsers{users: users})
	return nil
}

func (a *ActionListUsers) String() string {
	return "ListUsers"
}

type ActionSelectUser struct {
	user *User
}

func (a *ActionSelectUser) Do(ctx context.Context, u *UI) error {
	u.drawer.OnEvent(&EventSelectUser{user: a.user})
	return nil
}

func (a *ActionSelectUser) String() string {
	return "SelectUser"
}

type ActionCreateUser struct {
	username string
}

func (a *ActionCreateUser) Do(ctx context.Context, u *UI) error {
	user, err := u.CreateUser(ctx, a.username)
	if err != nil {
		return fmt.Errorf("CreateUser: %w", err)
	}
	u.drawer.OnEvent(&EventNewUser{user: user})
	return nil
}

func (a *ActionCreateUser) String() string {
	return "CreateUser"
}

type ActionFetchMessages struct {
	user *User
}

func (a *ActionFetchMessages) Do(ctx context.Context, u *UI) error {
	messages, err := a.user.FetchMessages(ctx)
	if err != nil {
		return fmt.Errorf("FetchMessages: %w", err)
	}
	u.drawer.OnEvent(&EventAddMessages{messages: messages})
	return nil
}

func (a *ActionFetchMessages) String() string {
	return "FetchMessages"
}

type ActionCreateConversation struct {
	localUser      *User
	remoteUsername string
}

func (a *ActionCreateConversation) Do(ctx context.Context, u *UI) error {
	_, err := a.localUser.GetPublicUser(ctx, a.remoteUsername)
	if err != nil {
		return fmt.Errorf("client.GetPublicUser: %w", err)
	}
	u.drawer.OnEvent(&EventNewConversation{conversation: &types.Conversation{
		LocalUser:  a.localUser.name,
		RemoteUser: a.remoteUsername,
	}})
	return nil
}

func (a *ActionCreateConversation) String() string {
	return "CreateConversation"
}

type ActionSelectConversation struct {
	conversation *types.Conversation
}

func (a *ActionSelectConversation) Do(ctx context.Context, u *UI) error {
	u.drawer.OnEvent(&EventSelectConversation{conversation: a.conversation})
	return nil
}

func (a *ActionSelectConversation) String() string {
	return "SelectConversation"
}

type ActionSendMessage struct {
	localUser      *User
	remoteUsername string
	plaintext      types.PlainText
}

func (a *ActionSendMessage) Do(ctx context.Context, u *UI) error {
	err := a.localUser.SendMessage(ctx, a.plaintext, a.remoteUsername)
	if err != nil {
		return fmt.Errorf("SendMessage(%#v): %w", a, err)
	}
	return nil
}

func (a *ActionSendMessage) String() string {
	return "SendMessage"
}

type ActionSwitchTab struct {
	tabIndex TabIndex
}

func (a *ActionSwitchTab) Do(ctx context.Context, u *UI) error {
	u.drawer.OnEvent(&EventSwitchTab{tabIndex: a.tabIndex})
	return nil
}

func (a *ActionSwitchTab) String() string {
	return "SwitchTab"
}
