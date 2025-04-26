package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/marc921/talk/internal/server/controller"
	"github.com/marc921/talk/internal/types/openapi"
)

type Authenticator struct {
	challengeSecretKey  []byte
	authSecretKey       []byte
	nonceLen            int
	challengeExpiration time.Duration
	authExpiration      time.Duration
}

func NewAuthenticator(
	challengeSecretKey []byte,
	authSecretKey []byte,
	nonceLen int,
	challengeExpiration time.Duration,
	authExpiration time.Duration,
) *Authenticator {
	return &Authenticator{
		challengeSecretKey:  challengeSecretKey,
		authSecretKey:       authSecretKey,
		nonceLen:            nonceLen,
		challengeExpiration: challengeExpiration,
		authExpiration:      authExpiration,
	}
}

// GenerateNonce generates a random nonce message
func (a *Authenticator) GenerateNonce() (string, error) {
	bytes := make([]byte, a.nonceLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateJWT generates a JWT with a nonce message as the payload
func (a *Authenticator) GenerateChallengeJWT(username openapi.Username, nonce string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   username,
		"nonce": nonce,
		"exp":   time.Now().Add(a.challengeExpiration).Unix(),
	})

	tokenString, err := token.SignedString(a.challengeSecretKey)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return tokenString, nil
}

func (a *Authenticator) GenerateAuthChallenge(username openapi.Username) (*openapi.AuthChallenge, error) {
	nonce, err := a.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("GenerateNonce: %w", err)
	}
	jwt, err := a.GenerateChallengeJWT(username, nonce)
	if err != nil {
		return nil, fmt.Errorf("GenerateJWT: %w", err)
	}
	return &openapi.AuthChallenge{
		Nonce: nonce,
		Token: jwt,
	}, nil
}

func (a *Authenticator) VerifyAuthChallenge(
	ctx context.Context,
	signedAuthChallenge *openapi.AuthChallengeSigned,
	controller *controller.ServerController,
) (openapi.Username, error) {
	token, err := jwt.Parse(
		signedAuthChallenge.Token,
		func(token *jwt.Token) (any, error) {
			return a.challengeSecretKey, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return "", fmt.Errorf("jwt.Parse: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return "", fmt.Errorf("missing expiration time")
	}
	if time.Now().After(expiration.Time) {
		return "", fmt.Errorf("token expired")
	}

	nonce, ok := claims["nonce"].(string)
	if !ok {
		return "", fmt.Errorf("missing nonce")
	}

	username, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing subject")
	}
	publicKey, err := controller.GetUserPublicKey(
		ctx,
		openapi.Username(username),
	)
	if err != nil {
		return "", fmt.Errorf("controller.GetUserPublicKey: %w", err)
	}

	nonceBytes, err := base64.URLEncoding.DecodeString(nonce)
	if err != nil {
		return "", fmt.Errorf("base64.URLEncoding.DecodeString: %w", err)
	}
	signedNonceBytes, err := base64.URLEncoding.DecodeString(signedAuthChallenge.SignedNonce)
	if err != nil {
		return "", fmt.Errorf("base64.URLEncoding.DecodeString: %w", err)
	}
	err = rsa.VerifyPKCS1v15(publicKey, 0, nonceBytes, signedNonceBytes)
	if err != nil {
		return "", fmt.Errorf("rsa.VerifyPKCS1v15: %w", err)
	}

	return username, nil
}

// GenerateJWT generates a JWT that authenticates the user
func (a *Authenticator) GenerateAuthJWT(username openapi.Username) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(a.authExpiration).Unix(),
	})

	tokenString, err := token.SignedString(a.authSecretKey)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}

	return tokenString, nil
}

func (a *Authenticator) VerifyAuthJWT(c echo.Context, username string) error {
	token, ok := c.Get("user").(*jwt.Token) // by default token is stored under `user` key
	if !ok {
		return fmt.Errorf("missing token")
	}

	claims, ok := token.Claims.(jwt.MapClaims) // by default claims is of type `jwt.MapClaims`
	if !ok {
		return fmt.Errorf("invalid claims")
	}

	sub := claims["sub"].(string)
	if sub == "" {
		return fmt.Errorf("missing subject")
	}

	if username != claims["sub"].(string) {
		return fmt.Errorf("wrong subject")
	}

	return nil
}
