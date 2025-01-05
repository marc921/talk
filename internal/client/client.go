package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"

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

func (c *Client) WebSocket(
	ctx context.Context,
	token string,
	readChan chan<- *openapi.Message,
	writeChan <-chan *openapi.Message,
) error {
	serverUrl, err := url.ParseRequestURI(UISingleton.config.Server.URL)
	if err != nil {
		return fmt.Errorf("url.ParseRequestURI: %w", err)
	}
	serverUrl.Scheme = "wss"
	serverUrl.Path = fmt.Sprintf("%s/ws/%s", serverUrl.Path, string(c.username))

	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	wsConn, _, err := websocket.DefaultDialer.DialContext(ctx, serverUrl.String(), header)
	if err != nil {
		return fmt.Errorf("websocket.DefaultDialer.Dial: %w", err)
	}
	defer wsConn.Close()

	wsConn.SetReadLimit(1 << 20) // 1 MiB	// TODO: chunk messages, match size with server 512

	errGrp, ctx := errgroup.WithContext(ctx)

	errGrp.Go(func() error {
		defer close(readChan)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				msgType, message, err := wsConn.ReadMessage()
				if err != nil {
					return fmt.Errorf("wsConn.ReadMessage: %w", err)
				}
				if msgType != websocket.TextMessage {
					// TODO: handle ping/pong messages ?
					return fmt.Errorf("unexpected message type: %d", msgType)
				}
				var msg *openapi.Message
				err = json.Unmarshal(message, &msg)
				if err != nil {
					return fmt.Errorf("json.Unmarshal: %w", err)
				}
				readChan <- msg
			}
		}
	})

	errGrp.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				err := wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					return fmt.Errorf("wsConn.WriteMessage(Close): %w", err)
				}
				return ctx.Err()
			case message, ok := <-writeChan:
				if !ok {
					return errors.New("writeChan closed")
				}
				msgBytes, err := json.Marshal(message)
				if err != nil {
					return fmt.Errorf("json.Marshal: %w", err)
				}
				err = wsConn.WriteMessage(websocket.TextMessage, msgBytes)
				if err != nil {
					return fmt.Errorf("wsConn.WriteMessage: %w", err)
				}
			}
		}
	})

	err = errGrp.Wait()
	if err != nil {
		return fmt.Errorf("errGrp.Wait: %w", err)
	}
	return nil
}
