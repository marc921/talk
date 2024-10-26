package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.L().Fatal("zap.NewDevelopment", zap.Error(err))
	}

	err = writeKeys("alice")
	if err != nil {
		logger.Fatal("writeKeys", zap.Error(err))
	}
	err = writeKeys("bob")
	if err != nil {
		logger.Fatal("writeKeys", zap.Error(err))
	}

	alice, err := NewUser(UserName("alice"))
	if err != nil {
		logger.Fatal("NewUser", zap.Error(err))
	}
	bob, err := NewUser(UserName("bob"))
	if err != nil {
		logger.Fatal("NewUser", zap.Error(err))
	}

	server := NewServer()
	err = server.AddUser(alice.name, &alice.key.PublicKey)
	if err != nil {
		logger.Fatal("AddUser", zap.Error(err))
	}
	err = server.AddUser(bob.name, &bob.key.PublicKey)
	if err != nil {
		logger.Fatal("AddUser", zap.Error(err))
	}

	err = alice.Send(PlainText("hello, bob"), bob.name, server)
	if err != nil {
		logger.Fatal("alice.Send", zap.Error(err))
	}

	err = alice.Send(PlainText("how are you?"), bob.name, server)
	if err != nil {
		logger.Fatal("alice.Send", zap.Error(err))
	}

	messages, err := bob.Receive(server)
	if err != nil {
		logger.Fatal("bob.Receive", zap.Error(err))
	}
	for _, message := range messages {
		fmt.Println(string(message))
	}
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
