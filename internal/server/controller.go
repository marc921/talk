package server

import (
	"crypto/rsa"
	"fmt"

	"github.com/marc921/talk/internal/cryptography"
	"github.com/marc921/talk/internal/types"
	"github.com/marc921/talk/internal/types/openapi"
	"go.uber.org/zap"
)

type ServerController struct {
	logger *zap.Logger
	// would be a database in a real application
	usersPublicKeys map[openapi.Username]*rsa.PublicKey
	inboxes         map[openapi.Username][]*openapi.Message
}

func NewServerController(logger *zap.Logger) *ServerController {
	return &ServerController{
		logger:          logger.With(zap.String("component", "controller")),
		usersPublicKeys: make(map[openapi.Username]*rsa.PublicKey),
		inboxes:         make(map[openapi.Username][]*openapi.Message),
	}
}

// AddUser adds a user to the server's database.
// If the user already exists, it returns true as the first return value.
func (s *ServerController) AddUser(username openapi.Username, publicKeyBytes []byte) (bool, error) {
	publicKey, err := cryptography.UnmarshalPublicKey(publicKeyBytes)
	if err != nil {
		return false, fmt.Errorf("cryptography.UnmarshalPublicKey: %w", err)
	}
	if existingPublicKey, ok := s.usersPublicKeys[username]; ok {
		if existingPublicKey.Equal(publicKey) {
			// idempotent: the user already exists with the same public key
			return true, nil
		}
		return true, types.ErrUserAlreadyExists
	}
	s.usersPublicKeys[username] = publicKey
	return false, nil
}

func (s *ServerController) GetUserPublicKey(username openapi.Username) (*rsa.PublicKey, bool) {
	publicKey, ok := s.usersPublicKeys[username]
	return publicKey, ok
}

func (s *ServerController) AddMessage(message *openapi.Message) {
	if _, ok := s.inboxes[message.Recipient]; !ok {
		s.inboxes[message.Recipient] = make([]*openapi.Message, 0)
	}
	s.inboxes[message.Recipient] = append(s.inboxes[message.Recipient], message)
}

func (s *ServerController) GetMessages(username openapi.Username) []*openapi.Message {
	messages, ok := s.inboxes[username]
	if !ok {
		return make([]*openapi.Message, 0)
	}
	s.inboxes[username] = make([]*openapi.Message, 0)
	return messages
}
