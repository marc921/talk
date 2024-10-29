package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path"
)

type Storage struct {
	HomeDir string
}

func NewStorage(homeDir string) *Storage {
	return &Storage{HomeDir: homeDir}
}

func (s *Storage) GetOrCreatePrivateKey(username string) (*rsa.PrivateKey, error) {
	privKeyPath := path.Join(s.HomeDir, "users", username, "private.pem")
	_, err := os.Stat(privKeyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = s.WriteKeys(username)
			if err != nil {
				return nil, fmt.Errorf("writeKeys: %w", err)
			}
		} else {
			return nil, fmt.Errorf("os.Stat: %w", err)
		}
	}

	privateKeyPEM, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}
	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	if privateKeyBlock == nil {
		return nil, fmt.Errorf("pem.Decode: no key found")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509.ParsePKCS1PrivateKey: %w", err)
	}
	return privateKey, nil
}

func (s *Storage) WriteKeys(username string) error {
	userDir := path.Join(s.HomeDir, "users", username)
	privKeyPath := path.Join(userDir, "private.pem")
	pubKeyPath := path.Join(userDir, "public.pem")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("rsa.GenerateKey: %w", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	err = os.MkdirAll(userDir, 0755)
	if err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	err = os.WriteFile(privKeyPath, privateKeyPEM, 0644)
	if err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}

	publicKey := &privateKey.PublicKey

	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(pubKeyPath, publicKeyPEM, 0644)
	if err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}

	return nil
}
