package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/marc921/talk/internal/cryptography"
	"github.com/marc921/talk/internal/types"
	"github.com/marc921/talk/internal/types/openapi"
	"go.uber.org/zap"
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
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Error: fmt.Sprintf("GenerateAuthChallenge: %v", err),
		})
	}
	return c.JSON(http.StatusOK, challenge)
}

func (a *API) PostAuth(c echo.Context) error {
	var signedAuthChallenge *openapi.AuthChallengeSigned
	if err := c.Bind(&signedAuthChallenge); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request").
			WithInternal(fmt.Errorf("c.Bind: %w", err))
	}

	username, err := a.Authenticator.VerifyAuthChallenge(signedAuthChallenge, a.Controller)
	if err != nil {
		a.logger.Warn("VerifyAuthChallenge", zap.Error(err))
		return c.JSON(http.StatusUnauthorized, openapi.ErrorResponse{
			Error: "unauthorized",
		})
	}

	authToken, err := a.Authenticator.GenerateAuthJWT(username)
	return c.JSON(http.StatusOK, openapi.JWT{
		Token: authToken,
	})
}

func (a *API) GetUser(c echo.Context) error {
	username := c.Param("username")

	publicKey, found := a.Controller.GetUserPublicKey(username)
	if !found {
		return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
			Error: "user not found",
		})
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

	alreadyExists, err := a.Controller.AddUser(req.Name, req.PublicKey)
	if err != nil {
		if errors.Is(err, types.ErrUserAlreadyExists) {
			return c.JSON(http.StatusConflict, openapi.ErrorResponse{
				Error: types.ErrUserAlreadyExists.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Error: err.Error(),
		})
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

	messages := a.Controller.GetMessages(username)

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

	a.Controller.AddMessage(message)

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
