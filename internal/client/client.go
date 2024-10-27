package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/marc921/talk/internal/cryptography"
	"github.com/marc921/talk/internal/types"
	"github.com/marc921/talk/internal/types/openapi"
)

type Client struct {
	openapiClient *openapi.ClientWithResponses
	username      openapi.Username
}

func NewClient(
	openapiClient *openapi.ClientWithResponses,
	username openapi.Username,
) *Client {
	return &Client{
		openapiClient: openapiClient,
		username:      username,
	}
}

func (c *Client) GetPublicUser(
	ctx context.Context,
	username openapi.Username,
) (*types.PublicUser, error) {
	resp, err := c.openapiClient.GetUsersUsernameWithResponse(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("GetUsersUsernameWithResponse: %w", err)
	}
	switch resp.HTTPResponse.StatusCode {
	case http.StatusOK:
		recipientKey, err := cryptography.UnmarshalPublicKey(resp.JSON200.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("cryptography.UnmarshalPublicKey: %w", err)
		}
		return &types.PublicUser{
			Name:      username,
			PublicKey: recipientKey,
		}, nil
	case http.StatusNotFound:
		return nil, errors.New(resp.JSON404.Error)
	default:
		return nil, fmt.Errorf("received unexpected status code: %d", resp.HTTPResponse.StatusCode)
	}
}

func (c *Client) RegisterUser(
	ctx context.Context,
	publicKey []byte,
) error {
	resp, err := c.openapiClient.PostUsersWithResponse(ctx, openapi.PublicUser{
		Name:      c.username,
		PublicKey: publicKey,
	})
	if err != nil {
		return fmt.Errorf("PostUsersWithResponse: %w", err)
	}
	switch resp.HTTPResponse.StatusCode {
	case http.StatusOK:
		// User already exists with the same public key
		return nil
	case http.StatusCreated:
		// User created
		return nil
	case http.StatusConflict:
		// User already exists with a different public key
		return errors.New(resp.JSON409.Error)
	default:
		return fmt.Errorf("received unexpected status code: %d", resp.HTTPResponse.StatusCode)
	}
}

func (c *Client) GetAuth(
	ctx context.Context,
) (*openapi.AuthChallenge, error) {
	resp, err := c.openapiClient.GetAuthUsernameWithResponse(ctx, c.username)
	if err != nil {
		return nil, fmt.Errorf("GetAuthUsernameWithResponse: %w", err)
	}
	switch resp.HTTPResponse.StatusCode {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusNotFound:
		return nil, errors.New(resp.JSON404.Error)
	default:
		return nil, fmt.Errorf("received unexpected status code: %d", resp.HTTPResponse.StatusCode)
	}
}

func (c *Client) PostAuth(
	ctx context.Context,
	challenge *openapi.AuthChallenge,
	signedNonce string,
) (*openapi.JWT, error) {
	resp, err := c.openapiClient.PostAuthUsernameWithResponse(ctx, c.username, openapi.PostAuthUsernameJSONRequestBody{
		SignedNonce: signedNonce,
		Token:       challenge.Token,
	})
	if err != nil {
		return nil, fmt.Errorf("PostAuthUsernameWithResponse: %w", err)
	}
	switch resp.HTTPResponse.StatusCode {
	case http.StatusOK:
		return resp.JSON200, nil
	case http.StatusUnauthorized:
		return nil, errors.New(resp.JSON401.Error)
	default:
		return nil, fmt.Errorf("received unexpected status code: %d", resp.HTTPResponse.StatusCode)
	}
}

func WithBearerToken(token string) openapi.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	}
}

func (c *Client) SendMessage(
	ctx context.Context,
	token string,
	message *openapi.Message,
) error {
	if message.Sender != c.username {
		return errors.New("cannot send message on behalf of another user")
	}
	if message.Recipient == c.username {
		return errors.New("cannot send message to self")
	}
	resp, err := c.openapiClient.PostMessagesUsernameWithResponse(
		ctx,
		c.username,
		*message,
		WithBearerToken(token),
	)
	if err != nil {
		return fmt.Errorf("PostMessagesUsernameWithResponse: %w", err)
	}
	switch resp.HTTPResponse.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return errors.New(resp.JSON401.Error)
	case http.StatusNotFound:
		return errors.New(resp.JSON404.Error)
	default:
		return fmt.Errorf("received unexpected status code: %d", resp.HTTPResponse.StatusCode)
	}
}

func (c *Client) GetMessages(
	ctx context.Context,
	token string,
) ([]openapi.Message, error) {
	resp, err := c.openapiClient.GetMessagesUsernameWithResponse(
		ctx,
		c.username,
		WithBearerToken(token),
	)
	if err != nil {
		return nil, fmt.Errorf("GetMessagesUsernameWithResponse: %w", err)
	}
	switch resp.HTTPResponse.StatusCode {
	case http.StatusOK:
		return *resp.JSON200, nil
	case http.StatusUnauthorized:
		return nil, errors.New(resp.JSON401.Error)
	case http.StatusNotFound:
		return nil, errors.New(resp.JSON404.Error)
	default:
		return nil, fmt.Errorf("received unexpected status code: %d", resp.HTTPResponse.StatusCode)
	}
}
