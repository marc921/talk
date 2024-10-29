package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"

	"github.com/marc921/talk/internal/cryptography"
	"github.com/marc921/talk/internal/types"
	"github.com/marc921/talk/internal/types/openapi"
)

type User struct {
	name      openapi.Username
	key       *rsa.PrivateKey
	authToken *string
	client    *Client
}

func NewUser(
	name openapi.Username,
	privKey *rsa.PrivateKey,
	client *Client,
) *User {
	return &User{
		name:   name,
		key:    privKey,
		client: client,
	}
}

func (u *User) Register(ctx context.Context) error {
	pubKeyBytes := cryptography.MarshalPublicKey(&u.key.PublicKey)
	err := u.client.RegisterUser(ctx, pubKeyBytes)
	if err != nil {

		return fmt.Errorf("client.RegisterUser: %w", err)
	}
	return nil
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

func (u *User) SendMessage(ctx context.Context, plaintext types.PlainText, recipientName openapi.Username) error {
	recipient, err := u.client.GetPublicUser(ctx, recipientName)
	if err != nil {
		return fmt.Errorf("client.GetPublicUser: %w", err)
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, recipient.PublicKey, plaintext)
	if err != nil {
		return fmt.Errorf("rsa.EncryptPKCS1v15: %w", err)
	}

	if u.authToken == nil {
		err = u.Authenticate(ctx)
		if err != nil {
			return fmt.Errorf("Authenticate: %w", err)
		}
	}

	err = u.client.SendMessage(ctx, *u.authToken, &openapi.Message{
		Sender:     u.name,
		Recipient:  recipientName,
		Ciphertext: ciphertext,
	})
	if err != nil {
		return fmt.Errorf("client.SendMessage: %w", err)
	}
	return nil
}

func (u *User) FetchMessages(ctx context.Context) ([]*types.PlainMessage, error) {
	if u.authToken == nil {
		err := u.Authenticate(ctx)
		if err != nil {
			return nil, fmt.Errorf("Authenticate: %w", err)
		}
	}

	encryptedMessages, err := u.client.GetMessages(ctx, *u.authToken)
	if err != nil {
		return nil, fmt.Errorf("client.GetMessages: %w", err)
	}
	var plainMessages []*types.PlainMessage
	for _, encryptedMessage := range encryptedMessages {
		plaintext, err := rsa.DecryptPKCS1v15(nil, u.key, []byte(encryptedMessage.Ciphertext))
		if err != nil {
			return nil, fmt.Errorf("rsa.DecryptPKCS1v15: %w", err)
		}
		plainMessages = append(plainMessages, &types.PlainMessage{
			Sender:    encryptedMessage.Sender,
			Recipient: encryptedMessage.Recipient,
			Plaintext: plaintext,
		})
	}
	return plainMessages, nil
}
