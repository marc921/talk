package server

import (
	"bytes"
	"context"
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/marc921/talk/internal/cryptography"
	"github.com/marc921/talk/internal/server/database/sqlcgen"
	"github.com/marc921/talk/internal/types"
	"github.com/marc921/talk/internal/types/openapi"
)

type ServerController struct {
	logger *zap.Logger
	db     *sql.DB
}

func NewServerController(
	logger *zap.Logger,
	db *sql.DB,
) *ServerController {
	return &ServerController{
		logger: logger.With(zap.String("component", "controller")),
		db:     db,
	}
}

// AddUser adds a user to the server's database.
// If the user already exists, it returns true as the first return value.
func (s *ServerController) AddUser(
	ctx context.Context,
	username openapi.Username,
	publicKeyBytes []byte,
) (bool, error) {
	// Validate key format
	_, err := cryptography.UnmarshalPublicKey(publicKeyBytes)
	if err != nil {
		return false, fmt.Errorf("cryptography.UnmarshalPublicKey: %w", err)
	}
	// Add user to the database
	queries := sqlcgen.New(s.db)
	user, err := queries.InsertUser(ctx, sqlcgen.InsertUserParams{
		Name:      username,
		PublicKey: publicKeyBytes,
	})
	if err != nil {
		return false, fmt.Errorf("queries.InsertUser: %w", err)
	}
	if !bytes.Equal(user.PublicKey, publicKeyBytes) {
		return true, types.ErrUserAlreadyExists
	}

	return user.Column1, nil
}

func (s *ServerController) GetUserPublicKey(
	ctx context.Context,
	username openapi.Username,
) (*rsa.PublicKey, error) {
	queries := sqlcgen.New(s.db)
	user, err := queries.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrNotFound
		}
		return nil, fmt.Errorf("queries.GetUser: %w", err)
	}
	publicKey, err := cryptography.UnmarshalPublicKey(user.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("cryptography.UnmarshalPublicKey: %w", err)
	}
	return publicKey, nil
}

func (s *ServerController) AddMessage(
	ctx context.Context,
	message *openapi.Message,
) error {
	queries := sqlcgen.New(s.db)
	_, err := queries.InsertMessage(ctx, sqlcgen.InsertMessageParams{
		Sender:       message.Sender,
		Recipient:    message.Recipient,
		CipherSymKey: message.CipherSymKey,
		Ciphertext:   message.Ciphertext,
	})
	if err != nil {
		return fmt.Errorf("queries.InsertMessage: %w", err)
	}

	return nil
}

func (s *ServerController) GetMessages(
	ctx context.Context,
	username openapi.Username,
) ([]*openapi.Message, error) {
	queries := sqlcgen.New(s.db)
	dbMessages, err := queries.GetUndeliveredMessages(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("queries.GetUndeliveredMessages: %w", err)
	}
	messages := make([]*openapi.Message, len(dbMessages))
	for i, dbMessage := range dbMessages {
		messages[i] = &openapi.Message{
			Sender:       dbMessage.Sender,
			Recipient:    dbMessage.Recipient,
			CipherSymKey: dbMessage.CipherSymKey,
			Ciphertext:   dbMessage.Ciphertext,
		}
	}

	// Mark messages as delivered
	for _, dbMessage := range dbMessages {
		err := queries.SetMessageDelivered(ctx, dbMessage.ID)
		if err != nil {
			return nil, fmt.Errorf("queries.SetMessageDelivered: %w", err)
		}
	}

	return messages, nil
}
