package main

import (
	"crypto/rsa"
	"fmt"
)

type PlainText []byte
type CipherText []byte
type UserName string

type Server struct {
	// would be a database in a real application
	usersPublicKeys map[UserName]*rsa.PublicKey
	inboxes         map[UserName][]CipherText
}

func NewServer() *Server {
	return &Server{
		usersPublicKeys: make(map[UserName]*rsa.PublicKey),
		inboxes:         make(map[UserName][]CipherText),
	}
}

func (s *Server) AddUser(username UserName, key *rsa.PublicKey) error {
	if _, ok := s.usersPublicKeys[username]; ok {
		return fmt.Errorf("user %s already exists", username)
	}
	s.usersPublicKeys[username] = key
	return nil
}

func (s *Server) GetUserPublicKey(username UserName) *rsa.PublicKey {
	return s.usersPublicKeys[username]
}

func (s *Server) AddMessage(recipient UserName, message CipherText) {
	if _, ok := s.inboxes[recipient]; !ok {
		s.inboxes[recipient] = make([]CipherText, 0)
	}
	s.inboxes[recipient] = append(s.inboxes[recipient], message)
}

func (s *Server) GetMessages(username UserName) []CipherText {
	messages, ok := s.inboxes[username]
	if !ok {
		return []CipherText{}
	}
	s.inboxes[username] = []CipherText{}
	return messages
}
