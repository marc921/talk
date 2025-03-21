package main

import (
	"context"
	"fmt"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	// Used to sign the challenge JWTs
	AuthChallengeSecretKey []byte `env:"AUTH_CHALLENGE_SECRET_KEY, required"`
	// Used to sign the auth tokens
	AuthTokenSecretKey []byte `env:"AUTH_TOKEN_SECRET_KEY, required"`
	TLS                bool   `env:"TLS, default=true"`
	DatabaseURL        string `env:"DATABASE_URL, required"`
}

func LoadConfig(ctx context.Context) (*Config, error) {
	cfg := new(Config)
	if err := envconfig.Process(ctx, cfg); err != nil {
		return nil, fmt.Errorf("envconfig.Process: %w", err)
	}
	return cfg, nil
}
