package server

import (
	"errors"
	"fmt"
	"image/color"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"

	"github.com/marc921/talk/internal/cryptography"
	"github.com/marc921/talk/internal/types"
	"github.com/marc921/talk/internal/types/openapi"
)

type API struct {
	logger        *zap.Logger
	Authenticator *Authenticator
	Controller    *ServerController
	WebsocketHub  *WebSocketHub
}

func NewAPI(
	logger *zap.Logger,
	authenticator *Authenticator,
	controller *ServerController,
	websocketHub *WebSocketHub,
) *API {
	return &API{
		logger:        logger.With(zap.String("component", "api")),
		Authenticator: authenticator,
		Controller:    controller,
		WebsocketHub:  websocketHub,
	}
}

func (a *API) GetAuth(c echo.Context) error {
	username := c.Param("username")
	challenge, err := a.Authenticator.GenerateAuthChallenge(openapi.Username(username))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "GenerateAuthChallenge").
			WithInternal(fmt.Errorf("Authenticator.GenerateAuthChallenge: %w", err))
	}
	return c.JSON(http.StatusOK, challenge)
}

func (a *API) PostAuth(c echo.Context) error {
	var signedAuthChallenge *openapi.AuthChallengeSigned
	if err := c.Bind(&signedAuthChallenge); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request").
			WithInternal(fmt.Errorf("c.Bind: %w", err))
	}

	username, err := a.Authenticator.VerifyAuthChallenge(
		c.Request().Context(),
		signedAuthChallenge,
		a.Controller,
	)
	if err != nil {
		a.logger.Warn("VerifyAuthChallenge", zap.Error(err))
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized").
			WithInternal(fmt.Errorf("Authenticator.VerifyAuthChallenge: %w", err))
	}

	authToken, err := a.Authenticator.GenerateAuthJWT(username)
	return c.JSON(http.StatusOK, openapi.JWT{
		Token: authToken,
	})
}

func (a *API) GetUser(c echo.Context) error {
	username := c.Param("username")

	publicKey, err := a.Controller.GetUserPublicKey(
		c.Request().Context(),
		username,
	)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, openapi.ErrorResponse{
				Error: "user not found",
			}).
				WithInternal(fmt.Errorf("Controller.GetUserPublicKey: %w", err))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user").
			WithInternal(fmt.Errorf("Controller.GetUserPublicKey: %w", err))
	}

	publicKeyPem := cryptography.MarshalPublicKey(publicKey)

	return c.JSON(http.StatusOK, openapi.PublicUser{
		Name:      username,
		PublicKey: publicKeyPem,
	})
}

func (a *API) AddUser(c echo.Context) error {
	var req *openapi.PublicUser
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request").
			WithInternal(fmt.Errorf("c.Bind: %w", err))
	}

	alreadyExists, err := a.Controller.AddUser(
		c.Request().Context(),
		req.Name,
		req.PublicKey,
	)
	if err != nil {
		if errors.Is(err, types.ErrUserAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, types.ErrUserAlreadyExists.Error()).
				WithInternal(fmt.Errorf("Controller.AddUser: %w", err))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add user").
			WithInternal(fmt.Errorf("Controller.AddUser: %w", err))
	}
	if alreadyExists {
		return c.JSON(http.StatusOK, nil)
	}

	return c.JSON(http.StatusCreated, nil)
}

func (a *API) GetMessages(c echo.Context) error {
	username := c.Param("username")

	err := a.Authenticator.VerifyAuthJWT(c, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized").
			WithInternal(fmt.Errorf("Authenticator.VerifyAuthJWT: %w", err))
	}

	messages, err := a.Controller.GetMessages(
		c.Request().Context(),
		username,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not get messages").
			WithInternal(fmt.Errorf("Controller.GetMessages: %w", err))
	}

	return c.JSON(http.StatusOK, messages)
}

func (a *API) AddMessage(c echo.Context) error {
	username := c.Param("username")

	err := a.Authenticator.VerifyAuthJWT(c, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized").
			WithInternal(fmt.Errorf("Authenticator.VerifyAuthJWT: %w", err))
	}

	var message *openapi.Message
	if err := c.Bind(&message); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request").
			WithInternal(fmt.Errorf("c.Bind: %w", err))
	}

	if username != message.Sender {
		return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
	}

	err = a.Controller.AddMessage(
		c.Request().Context(),
		message,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not add message").
			WithInternal(fmt.Errorf("Controller.AddMessage: %w", err))
	}

	return c.JSON(http.StatusCreated, nil)
}

// serveWs handles websocket requests from the peer.
func (a *API) RegisterWebsocketClient(c echo.Context) error {
	username := c.Param("username")

	err := a.Authenticator.VerifyAuthJWT(c, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized").
			WithInternal(fmt.Errorf("Authenticator.VerifyAuthJWT: %w", err))
	}

	err = a.WebsocketHub.RegisterClient(c, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).
			WithInternal(fmt.Errorf("WebsocketHub.RegisterClient: %w", err))
	}

	return nil
}

type GenerateQRCodeRequest struct {
	Content string `query:"content"`
	Size    int    `query:"size"`
	Color   string `query:"color"`
}

func (a *API) GenerateQRCode(c echo.Context) error {
	var req GenerateQRCodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request").
			WithInternal(fmt.Errorf("c.Bind: %w", err))
	}
	if req.Size == 0 {
		req.Size = 256
	}

	q, err := qrcode.New(req.Content, qrcode.Medium)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create QR code").
			WithInternal(fmt.Errorf("qrcode.New: %w", err))
	}

	switch req.Color {
	case "red":
		q.ForegroundColor = color.RGBA{255, 0, 0, 255}
	case "green":
		q.ForegroundColor = color.RGBA{0, 128, 0, 255}
	case "blue":
		q.ForegroundColor = color.RGBA{0, 0, 255, 255}
	case "purple":
		q.ForegroundColor = color.RGBA{128, 0, 128, 255}
	case "orange":
		q.ForegroundColor = color.RGBA{255, 165, 0, 255}
	default:
		// black
		q.ForegroundColor = color.RGBA{0, 0, 0, 255}
	}

	// Generate QR code
	qr, err := q.PNG(req.Size)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate QR code").
			WithInternal(fmt.Errorf("qrcode.Encode: %w", err))
	}

	return c.Blob(http.StatusOK, "image/png", qr)
}
