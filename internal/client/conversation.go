package client

import (
	"github.com/marc921/talk/internal/client/database/sqlcgen"
)

type Conversation struct {
	dbConv   *sqlcgen.Conversation
	messages []*sqlcgen.Message
}

func NewConversation(
	dbConv *sqlcgen.Conversation,
) *Conversation {
	return &Conversation{
		dbConv: dbConv,
	}
}
