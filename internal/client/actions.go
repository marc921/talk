package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/marc921/talk/internal/client/database/sqlcgen"
	"github.com/marc921/talk/internal/cryptography"
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

type ActionZero struct{}

func (a *ActionZero) Do(ctx context.Context, u *UI) error {
	return nil
}

func (a *ActionZero) String() string {
	return "Zero"
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
	return "SetMode"
}

type ActionListUsers struct{}

func (a *ActionListUsers) Do(ctx context.Context, u *UI) error {
	queries := sqlcgen.New(u.db)
	localUsers, err := queries.ListLocalUsers(ctx)
	if err != nil {
		return fmt.Errorf("queries.ListLocalUsers: %w", err)
	}
	users := make([]*User, 0, len(localUsers))
	for _, localUser := range localUsers {
		localUser := localUser
		user, err := NewUser(localUser, UISingleton.openapiClient, UISingleton.db)
		if err != nil {
			return fmt.Errorf("NewUser: %w", err)
		}
		err = user.RegisterWebSocket(ctx)
		if err != nil {
			return fmt.Errorf("user.RegisterWebSocket: %w", err)
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
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txQueries := sqlcgen.New(u.db).WithTx(tx)

	// Create user locally
	privKey, err := cryptography.GenerateKey()
	if err != nil {
		return fmt.Errorf("cryptography.GenerateKey: %w", err)
	}
	privKeyBytes := cryptography.MarshalPrivateKey(privKey)

	localUser, err := txQueries.InsertLocalUser(ctx, sqlcgen.InsertLocalUserParams{
		Name:       a.username,
		PrivateKey: privKeyBytes,
	})
	if err != nil {
		return fmt.Errorf("queries.InsertLocalUser: %w", err)
	}

	user, err := NewUser(localUser, UISingleton.openapiClient, UISingleton.db)
	if err != nil {
		return fmt.Errorf("NewUser: %w", err)
	}

	// Attempt to register user on distant server, fail if already exists
	pubKeyBytes := cryptography.MarshalPublicKey(&privKey.PublicKey)
	err = user.client.RegisterUser(ctx, pubKeyBytes)
	if err != nil {
		return fmt.Errorf("client.RegisterUser: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	err = user.RegisterWebSocket(ctx)
	if err != nil {
		return fmt.Errorf("user.RegisterWebSocket: %w", err)
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
	_, err := a.user.FetchMessages(ctx)
	if err != nil {
		return fmt.Errorf("user.FetchMessages: %w", err)
	}
	u.drawer.OnEvent(&EventUpdateUser{user: a.user})
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
	err := a.localUser.CreateConversation(ctx, a.remoteUsername)
	if err != nil {
		return fmt.Errorf("localUser.CreateConversation: %w", err)
	}
	u.drawer.OnEvent(&EventUpdateUser{user: a.localUser})
	return nil
}

func (a *ActionCreateConversation) String() string {
	return "CreateConversation"
}

type ActionSelectConversation struct {
	conversation *Conversation
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
