package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type User struct {
	name UserName
	key  *rsa.PrivateKey
}

func NewUser(name UserName) (*User, error) {
	privKey, err := loadPrivateKey(string(name))
	if err != nil {
		return nil, fmt.Errorf("loadPrivateKey: %w", err)
	}

	return &User{
		name: name,
		key:  privKey,
	}, nil
}

func loadPrivateKey(dir string) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := os.ReadFile(dir + "/private.pem")
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

func (u *User) Send(message PlainText, recipient UserName, server *Server) error {
	recipientKey := server.GetUserPublicKey(recipient)
	if recipientKey == nil {
		return fmt.Errorf("recipient %q not found", recipient)
	}
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, recipientKey, message)
	if err != nil {
		return fmt.Errorf("rsa.EncryptPKCS1v15: %w", err)
	}
	server.AddMessage(recipient, ciphertext)
	return nil
}

func (u *User) Receive(server *Server) ([]PlainText, error) {
	var messages []PlainText
	for _, ciphertext := range server.GetMessages(u.name) {
		message, err := rsa.DecryptPKCS1v15(rand.Reader, u.key, ciphertext)
		if err != nil {
			return nil, fmt.Errorf("rsa.DecryptPKCS1v15: %w", err)
		}
		messages = append(messages, message)
	}
	return messages, nil
}
