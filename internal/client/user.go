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
	name          openapi.Username
	db            *sql.DB
	key           *rsa.PrivateKey
	client        *Client
	authToken     *string
	conversations map[openapi.Username]*Conversation
}

func NewUser(
	localUser *sqlcgen.LocalUser,
	db *sql.DB,
	openapiClient *openapi.ClientWithResponses,
) (*User, error) {
	privKey, err := cryptography.UnmarshalPrivateKey(localUser.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("cryptography.UnmarshalPrivateKey: %w", err)
	}
	user := &User{
		name:          localUser.Name,
		db:            db,
		key:           privKey,
		client:        NewClient(openapiClient, localUser.Name),
		conversations: make(map[openapi.Username]*Conversation),
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

func (u *User) GetPublicUser(ctx context.Context, name openapi.Username) (*types.PublicUser, error) {
	queries := sqlcgen.New(u.db)
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
	queries := sqlcgen.New(u.db)
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
	queries := sqlcgen.New(u.db)
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

	recipient, err := u.GetPublicUser(ctx, recipientName)
	if err != nil {
		return fmt.Errorf("GetPublicUser: %w", err)
	}

	// Encrypt plaintext with new unique symmetric key
	symKey, err := cryptography.GenerateAESKey()
	if err != nil {
		return fmt.Errorf("cryptography.GenerateAESKey: %w", err)
	}
	cipher, err := cryptography.NewAESCipher(symKey)
	if err != nil {
		return fmt.Errorf("cryptography.NewAESCipher: %w", err)
	}
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("cipher.Encrypt: %w", err)
	}

	// Encrypt symmetric key with recipient's public key
	cipheredSymKey, err := rsa.EncryptPKCS1v15(rand.Reader, recipient.PublicKey, symKey)
	if err != nil {
		return fmt.Errorf("rsa.EncryptPKCS1v15: %w", err)
	}

	if u.authToken == nil {
		err := u.Authenticate(ctx)
		if err != nil {
			return fmt.Errorf("Authenticate: %w", err)
		}
	}

	// Start transaction to insert message in database and send it to remote user, rollback if any error occurs
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txQueries := sqlcgen.New(u.db).WithTx(tx)

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
	err = u.client.SendMessage(ctx, *u.authToken, &openapi.Message{
		Sender:       u.name,
		Recipient:    recipientName,
		CipherSymKey: cipheredSymKey,
		Ciphertext:   ciphertext,
	})
	if err != nil {
		return fmt.Errorf("client.SendMessage: %w", err)
	}

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

	queries := sqlcgen.New(u.db)
	for _, message := range messages {
		conv, ok := u.conversations[message.Sender]
		if !ok {
			err = u.CreateConversation(ctx, message.Sender)
			if err != nil {
				return fmt.Errorf("CreateConversation: %w", err)
			}
			conv = u.conversations[message.Sender]
		}
		// TODO: sign and verify messages
		// sender, err := u.GetPublicUser(ctx, message.Sender)
		// if err != nil {
		// 	return fmt.Errorf("GetPublicUser: %w", err)
		// }
		// Decrypt symmetric key with private key
		symKey, err := rsa.DecryptPKCS1v15(rand.Reader, u.key, message.CipherSymKey)
		if err != nil {
			return fmt.Errorf("rsa.DecryptPKCS1v15: %w", err)
		}
		cipher, err := cryptography.NewAESCipher(symKey)
		if err != nil {
			return fmt.Errorf("cryptography.NewAESCipher: %w", err)
		}
		plaintext, err := cipher.Decrypt(message.Ciphertext)
		if err != nil {
			return fmt.Errorf("cipher.Decrypt: %w", err)
		}
		// Insert message in database
		dbMessage, err := queries.InsertMessage(ctx, sqlcgen.InsertMessageParams{
			ConversationID: conv.dbConv.ID,
			Sender:         message.Sender,
			Receiver:       message.Recipient,
			Content:        plaintext,
		})
		if err != nil {
			return fmt.Errorf("queries.InsertMessage: %w", err)
		}
		conv.messages = append(conv.messages, dbMessage)

	}
	return nil
}
