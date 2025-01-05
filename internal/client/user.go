package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/marc921/talk/internal/client/database/sqlcgen"
	"github.com/marc921/talk/internal/cryptography"
	"github.com/marc921/talk/internal/types"
	"github.com/marc921/talk/internal/types/openapi"
)

type User struct {
	name             openapi.Username
	key              *rsa.PrivateKey
	client           *Client
	authToken        *string
	conversations    map[openapi.Username]*Conversation
	inboundMessages  chan *openapi.Message
	outboundMessages chan *openapi.Message
}

func NewUser(
	localUser *sqlcgen.LocalUser,
) (*User, error) {
	privKey, err := cryptography.UnmarshalPrivateKey(localUser.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("cryptography.UnmarshalPrivateKey: %w", err)
	}
	user := &User{
		name:             localUser.Name,
		key:              privKey,
		client:           NewClient(UISingleton.openapiClient, localUser.Name),
		conversations:    make(map[openapi.Username]*Conversation),
		inboundMessages:  make(chan *openapi.Message),
		outboundMessages: make(chan *openapi.Message),
	}
	return user, nil
}

func (u *User) Authenticate(ctx context.Context) error {
	// Get the nonce from the server
	challenge, err := u.client.GetAuth(ctx)
	if err != nil {
		return fmt.Errorf("GetAuth: %w", err)
	}
	nonceBytes, err := base64.URLEncoding.DecodeString(challenge.Nonce)
	if err != nil {
		return fmt.Errorf("base64.URLEncoding.DecodeString: %w", err)
	}

	// Sign the nonce with the user's private key
	signedNonceBytes, err := rsa.SignPKCS1v15(rand.Reader, u.key, 0, nonceBytes)
	if err != nil {
		return fmt.Errorf("rsa.SignPKCS1v15: %w", err)
	}
	signedNonce := base64.URLEncoding.EncodeToString(signedNonceBytes)

	// Send the signed nonce to the server
	authResp, err := u.client.PostAuth(ctx, challenge, signedNonce)
	if err != nil {
		return fmt.Errorf("client.PostAuth: %w", err)
	}
	u.authToken = &authResp.Token
	return nil
}

func (u *User) RegisterWebSocket(ctx context.Context) error {
	if u.authToken == nil {
		err := u.Authenticate(ctx)
		if err != nil {
			return fmt.Errorf("Authenticate: %w", err)
		}
	}
	go func() {
		err := u.client.WebSocket(ctx, *u.authToken, u.inboundMessages, u.outboundMessages)
		if err != nil {
			UISingleton.actions <- &ActionSetError{err: fmt.Errorf("client.WebSocket: %w", err)}
		}
	}()

	go func() {
		queries := sqlcgen.New(UISingleton.db)
		for message := range u.inboundMessages {
			err := u.receiveMessage(ctx, queries, *message)
			if err != nil {
				UISingleton.actions <- &ActionSetError{err: fmt.Errorf("receiveMessage: %w", err)}
			}
			UISingleton.drawer.Draw()
		}
	}()
	return nil
}

func (u *User) GetPublicUser(ctx context.Context, name openapi.Username) (*types.PublicUser, error) {
	queries := sqlcgen.New(UISingleton.db)
	// Check if the public user is already in the database
	publicUser, err := queries.GetPublicUserByName(ctx, name)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("queries.GetPublicUserByName: %w", err)
		}

		// If the user is not in the database, fetch it from the server
		resp, err := u.client.GetPublicUser(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("client.GetPublicUser: %w", err)
		}
		// Insert the public user into the database
		publicUser, err = queries.InsertPublicUser(ctx, sqlcgen.InsertPublicUserParams{
			Name:      resp.Name,
			PublicKey: cryptography.MarshalPublicKey(resp.PublicKey),
		})

	}

	pubKey, err := cryptography.UnmarshalPublicKey(publicUser.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("cryptography.UnmarshalPublicKey: %w", err)
	}
	return &types.PublicUser{
		Name:      publicUser.Name,
		PublicKey: pubKey,
	}, nil
}

// ListConversations fetches the conversations from the database along with their messages
// and stores them in the user's local cache
func (u *User) FetchConversationsFromDB(ctx context.Context) error {
	queries := sqlcgen.New(UISingleton.db)
	dbConvs, err := queries.ListConversations(ctx, u.name)
	if err != nil {
		return fmt.Errorf("queries.ListConversations: %w", err)
	}
	for _, dbConv := range dbConvs {
		dbConv := dbConv
		u.conversations[dbConv.RemoteUserName] = NewConversation(dbConv)
		dbMessages, err := queries.ListMessages(ctx, dbConv.ID)
		if err != nil {
			return fmt.Errorf("queries.ListMessages: %w", err)
		}
		u.conversations[dbConv.RemoteUserName].messages = dbMessages
	}
	return nil
}

func (u *User) CreateConversation(ctx context.Context, remoteUsername openapi.Username) error {
	queries := sqlcgen.New(UISingleton.db)
	dbConv, err := queries.InsertConversation(ctx, sqlcgen.InsertConversationParams{
		LocalUserName:  u.name,
		RemoteUserName: remoteUsername,
	})
	if err != nil {
		return fmt.Errorf("queries.InsertConversation: %w", err)
	}
	u.conversations[remoteUsername] = NewConversation(dbConv)
	return nil
}

func (u *User) SendMessage(ctx context.Context, plaintext types.PlainText, recipientName openapi.Username) error {
	conversation, ok := u.conversations[recipientName]
	if !ok {
		return fmt.Errorf("conversation not found")
	}

	encryptedMsg, err := u.encryptMessage(ctx, plaintext, recipientName)
	if err != nil {
		return fmt.Errorf("encryptMessage: %w", err)
	}

	if u.authToken == nil {
		err := u.Authenticate(ctx)
		if err != nil {
			return fmt.Errorf("Authenticate: %w", err)
		}
	}

	// Start transaction to insert message in database and send it to remote user, rollback if any error occurs
	tx, err := UISingleton.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txQueries := sqlcgen.New(UISingleton.db).WithTx(tx)

	// Insert message in database
	dbMessage, err := txQueries.InsertMessage(ctx, sqlcgen.InsertMessageParams{
		ConversationID: conversation.dbConv.ID,
		Sender:         u.name,
		Receiver:       recipientName,
		Content:        plaintext,
	})
	if err != nil {
		return fmt.Errorf("txQueries.InsertMessage: %w", err)
	}

	// Add message to conversation local cache
	conversation.messages = append(conversation.messages, dbMessage)

	// Send message to remote user
	// err = u.client.SendMessage(ctx, *u.authToken, encryptedMsg)
	// if err != nil {
	// 	return fmt.Errorf("client.SendMessage: %w", err)
	// }
	u.outboundMessages <- encryptedMsg

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("tx.Commit: %w", err)
	}

	return nil
}

func (u *User) FetchMessages(ctx context.Context) error {
	// Fetch conversations and messages from the database
	err := u.FetchConversationsFromDB(ctx)
	if err != nil {
		return fmt.Errorf("ListConversations: %w", err)
	}

	// Fetch additional messages from the server
	if u.authToken == nil {
		err := u.Authenticate(ctx)
		if err != nil {
			return fmt.Errorf("Authenticate: %w", err)
		}
	}

	messages, err := u.client.GetMessages(ctx, *u.authToken)
	if err != nil {
		return fmt.Errorf("client.GetMessages: %w", err)
	}

	queries := sqlcgen.New(UISingleton.db)
	for _, message := range messages {
		err := u.receiveMessage(ctx, queries, message)
		if err != nil {
			return fmt.Errorf("receiveMessage: %w", err)
		}
	}
	return nil
}

func (u *User) receiveMessage(
	ctx context.Context,
	queries *sqlcgen.Queries,
	message openapi.Message,
) error {
	conv, ok := u.conversations[message.Sender]
	if !ok {
		err := u.CreateConversation(ctx, message.Sender)
		if err != nil {
			return fmt.Errorf("CreateConversation: %w", err)
		}
		conv = u.conversations[message.Sender]
	}

	decryptedMsg, err := u.decryptMessage(message)
	if err != nil {
		return fmt.Errorf("decryptMessage: %w", err)
	}
	decryptedMsg.ConversationID = conv.dbConv.ID

	// Insert message in database
	dbMessage, err := queries.InsertMessage(ctx, *decryptedMsg)
	if err != nil {
		return fmt.Errorf("queries.InsertMessage: %w", err)
	}
	conv.messages = append(conv.messages, dbMessage)
	return nil
}

func (u *User) decryptMessage(message openapi.Message) (*sqlcgen.InsertMessageParams, error) {
	// TODO: sign and verify messages
	// sender, err := u.GetPublicUser(ctx, message.Sender)
	// if err != nil {
	// 	return fmt.Errorf("GetPublicUser: %w", err)
	// }
	// Decrypt symmetric key with private key
	symKey, err := rsa.DecryptPKCS1v15(rand.Reader, u.key, message.CipherSymKey)
	if err != nil {
		return nil, fmt.Errorf("rsa.DecryptPKCS1v15: %w", err)
	}
	cipher, err := cryptography.NewAESCipher(symKey)
	if err != nil {
		return nil, fmt.Errorf("cryptography.NewAESCipher: %w", err)
	}
	plaintext, err := cipher.Decrypt(message.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("cipher.Decrypt: %w", err)
	}

	return &sqlcgen.InsertMessageParams{
		Sender:   message.Sender,
		Receiver: message.Recipient,
		Content:  plaintext,
	}, nil
}

func (u *User) encryptMessage(
	ctx context.Context,
	plaintext types.PlainText,
	recipientName openapi.Username,
) (*openapi.Message, error) {
	recipient, err := u.GetPublicUser(ctx, recipientName)
	if err != nil {
		return nil, fmt.Errorf("GetPublicUser: %w", err)
	}

	// Encrypt plaintext with new unique symmetric key
	symKey, err := cryptography.GenerateAESKey()
	if err != nil {
		return nil, fmt.Errorf("cryptography.GenerateAESKey: %w", err)
	}
	cipher, err := cryptography.NewAESCipher(symKey)
	if err != nil {
		return nil, fmt.Errorf("cryptography.NewAESCipher: %w", err)
	}
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return nil, fmt.Errorf("cipher.Encrypt: %w", err)
	}

	// Encrypt symmetric key with recipient's public key
	cipheredSymKey, err := rsa.EncryptPKCS1v15(rand.Reader, recipient.PublicKey, symKey)
	if err != nil {
		return nil, fmt.Errorf("rsa.EncryptPKCS1v15: %w", err)
	}

	return &openapi.Message{
		Sender:       u.name,
		Recipient:    recipientName,
		CipherSymKey: cipheredSymKey,
		Ciphertext:   ciphertext,
	}, nil
}
