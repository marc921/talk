package client

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/marc921/talk/internal/client/database/sqlcgen"
	"github.com/marc921/talk/internal/cryptography"
	"github.com/marc921/talk/internal/types/openapi"
)

type Controller struct {
	openapiClient *openapi.ClientWithResponses
	db            *sql.DB
}

func NewController(
	openapiClient *openapi.ClientWithResponses,
	db *sql.DB,
) *Controller {
	return &Controller{
		openapiClient: openapiClient,
		db:            db,
	}
}

func (c *Controller) CreateUser(
	ctx context.Context,
	username string,
) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txQueries := sqlcgen.New(c.db).WithTx(tx)

	// Create user locally
	privKey, err := cryptography.GenerateKey()
	if err != nil {
		return fmt.Errorf("cryptography.GenerateKey: %w", err)
	}
	privKeyBytes := cryptography.MarshalPrivateKey(privKey)

	localUser, err := txQueries.InsertLocalUser(ctx, sqlcgen.InsertLocalUserParams{
		Name:       username,
		PrivateKey: privKeyBytes,
	})
	if err != nil {
		return fmt.Errorf("queries.InsertLocalUser: %w", err)
	}

	user, err := NewUser(localUser, c.openapiClient, c.db)
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

	return nil
}

func (c *Controller) GetUser(
	ctx context.Context,
	username openapi.Username,
) (*User, error) {
	queries := sqlcgen.New(c.db)
	localUser, err := queries.GetLocalUserByName(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("GetLocalUserByName: %w", err)
	}
	user, err := NewUser(localUser, c.openapiClient, c.db)
	if err != nil {
		return nil, fmt.Errorf("NewUser: %w", err)
	}

	return user, nil
}
