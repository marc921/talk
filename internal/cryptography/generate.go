package cryptography

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

func GenerateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("rsa.GenerateKey: %w", err)
	}
	return privateKey, nil
}
