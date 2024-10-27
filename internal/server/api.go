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
}

func NewAPI(
	logger *zap.Logger,
	authenticator *Authenticator,
	controller *ServerController,
) *API {
	return &API{
		logger:        logger,
		Authenticator: authenticator,
		Controller:    controller,
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
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Error: fmt.Sprintf("invalid request: %v", err),
		})
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
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Error: fmt.Sprintf("invalid request: %v", err),
		})
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
		return c.JSON(http.StatusUnauthorized, openapi.ErrorResponse{
			Error: fmt.Sprintf("unauthorized: %v", err),
		})
	}

	messages := a.Controller.GetMessages(username)

	return c.JSON(http.StatusOK, messages)
}

func (a *API) AddMessage(c echo.Context) error {
	username := c.Param("username")

	err := a.Authenticator.VerifyAuthJWT(c, username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, openapi.ErrorResponse{
			Error: fmt.Sprintf("unauthorized: %v", err),
		})
	}

	var message *openapi.Message
	if err := c.Bind(&message); err != nil {
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Error: fmt.Sprintf("invalid request: %v", err),
		})
	}

	if username != message.Sender {
		return c.JSON(http.StatusForbidden, openapi.ErrorResponse{
			Error: "forbidden",
		})
	}

	a.Controller.AddMessage(message)

	return c.JSON(http.StatusCreated, nil)
}
