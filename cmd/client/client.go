package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/marc921/talk/internal/client"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.L().Fatal("zap.NewDevelopment", zap.Error(err))
	}

	config, err := LoadConfig(ctx)
	if err != nil {
		if errors.Is(err, ErrAbortedByUser) {
			return
		}
		logger.Fatal("LoadConfig", zap.Error(err))
	}

	ui, err := client.NewUI(config.Server.URL)
	if err != nil {
		logger.Fatal("NewUI", zap.Error(err))
	}
	defer ui.Close()

	err = ui.Run(ctx)
	if err != nil {
		logger.Fatal("ui.Run", zap.Error(err))
	}
}

func loadPrivateKey(username string) (*rsa.PrivateKey, error) {
	_, err := os.Stat(username + "/private.pem")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = writeKeys(username)
			if err != nil {
				return nil, fmt.Errorf("writeKeys: %w", err)
			}
		} else {
			return nil, fmt.Errorf("os.Stat: %w", err)
		}
	}

	privateKeyPEM, err := os.ReadFile(username + "/private.pem")
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

func writeKeys(username string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("rsa.GenerateKey: %w", err)
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	err = os.MkdirAll(username, 0755)
	if err != nil {
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	err = os.WriteFile(username+"/private.pem", privateKeyPEM, 0644)
	if err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}

	publicKey := &privateKey.PublicKey

	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	err = os.WriteFile(username+"/public.pem", publicKeyPEM, 0644)
	if err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}

	return nil
}

// alicePrivKey, err := loadPrivateKey("alice")
// if err != nil {
// 	logger.Fatal("loadPrivateKey", zap.Error(err))
// }

// alice, err := client.NewUser(
// 	openapi.Username("alice"),
// 	alicePrivKey,
// 	client.NewClient(openapiClient, openapi.Username("alice")),
// )
// if err != nil {
// 	logger.Fatal("NewUser", zap.Error(err))
// }

// bobPrivKey, err := loadPrivateKey("bob")
// if err != nil {
// 	logger.Fatal("loadPrivateKey", zap.Error(err))
// }

// bob, err := client.NewUser(
// 	openapi.Username("bob"),
// 	bobPrivKey,
// 	client.NewClient(openapiClient, openapi.Username("bob")),
// )
// if err != nil {
// 	logger.Fatal("NewUser", zap.Error(err))
// }

// err = alice.Register(ctx)
// if err != nil {
// 	logger.Fatal("alice.Register", zap.Error(err))
// }

// err = bob.Register(ctx)
// if err != nil {
// 	logger.Fatal("bob.Register", zap.Error(err))
// }

// err = alice.SendMessage(
// 	ctx,
// 	types.PlainText("hello, bob"),
// 	openapi.Username("bob"),
// )
// if err != nil {
// 	logger.Fatal("alice.SendMessage", zap.Error(err))
// }

// messages, err := bob.FetchMessages(ctx)
// if err != nil {
// 	logger.Fatal("bob.FetchMessages", zap.Error(err))
// }
// for _, message := range messages {
// 	fmt.Printf("%s -> %s: %q\n", message.Sender, message.Recipient, string(message.Plaintext))
// }
