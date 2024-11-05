package cryptography

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func GenerateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("rsa.GenerateKey: %w", err)
	}
	return privateKey, nil
}

func MarshalPublicKey(publicKey *rsa.PublicKey) []byte {
	publicKeyBytes := x509.MarshalPKCS1PublicKey(publicKey)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return publicKeyPEM
}

func UnmarshalPublicKey(pemBytes []byte) (*rsa.PublicKey, error) {
	publicKeyBlock, _ := pem.Decode(pemBytes)
	if publicKeyBlock == nil {
		return nil, fmt.Errorf("pem.Decode: no key found")
	}
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509.ParsePKCS1PublicKey: %w", err)
	}
	return publicKey, nil
}

func MarshalPrivateKey(privateKey *rsa.PrivateKey) []byte {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return privateKeyPEM
}

func UnmarshalPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	privateKeyBlock, _ := pem.Decode(pemBytes)
	if privateKeyBlock == nil {
		return nil, fmt.Errorf("pem.Decode: no key found")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("x509.ParsePKCS1PrivateKey: %w", err)
	}
	return privateKey, nil
}
