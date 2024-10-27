package types

import (
	"crypto/rsa"
	"errors"

	"github.com/marc921/talk/internal/types/openapi"
)

type PlainText = []byte

var ErrUserAlreadyExists = errors.New("user already exists")

type PublicUser struct {
	Name      openapi.Username
	PublicKey *rsa.PublicKey
}

type PlainMessage struct {
	Sender    openapi.Username
	Recipient openapi.Username
	Plaintext PlainText
}