package client

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/marc921/talk/internal/cryptography"
)

type Storage struct {
	HomeDir string
}

func NewStorage(homeDir string) *Storage {
	return &Storage{HomeDir: homeDir}
}

func (s *Storage) ListUsers() ([]string, error) {
	usersDir := path.Join(s.HomeDir, "users")
	dirEntries, err := os.ReadDir(usersDir)
	if err != nil {
		return nil, fmt.Errorf("os.ReadDir: %w", err)
	}
	var users []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			users = append(users, entry.Name())
		}
	}
	return users, nil
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

	privateKey, err := cryptography.UnmarshalPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("cryptography.UnmarshalPrivateKey: %w", err)
	}
	return privateKey, nil
}

func (s *Storage) WriteKeys(username string) error {
	userDir := path.Join(s.HomeDir, "users", username)
	privKeyPath := path.Join(userDir, "private.pem")
	pubKeyPath := path.Join(userDir, "public.pem")

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("rsa.GenerateKey: %w", err)
	}

	privateKeyBytes := cryptography.MarshalPrivateKey(privateKey)

	err = os.MkdirAll(userDir, 0755)
	if err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	err = os.WriteFile(privKeyPath, privateKeyBytes, 0644)
	if err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}

	publicKey := &privateKey.PublicKey

	publicKeyBytes := cryptography.MarshalPublicKey(publicKey)

	err = os.WriteFile(pubKeyPath, publicKeyBytes, 0644)
	if err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}

	return nil
}
