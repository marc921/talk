package types

import (
	"crypto/rsa"
	"errors"
	"time"

	"github.com/marc921/talk/internal/types/openapi"
)

type PlainText = []byte

var ErrUserAlreadyExists = errors.New("user already exists")

type PublicUser struct {
	Name      openapi.Username
	PublicKey *rsa.PublicKey
}

type PlainMessage struct {
	From        openapi.Username
	To          openapi.Username
	Plaintext   PlainText
	SentAt      *time.Time
	DeliveredAt *time.Time
	ReadAt      *time.Time
}
