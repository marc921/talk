// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package sqlcgen

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Message struct {
	ID           pgtype.UUID
	Sender       string
	Recipient    string
	CipherSymKey []byte
	Ciphertext   []byte
	SentAt       pgtype.Timestamptz
	DeliveredAt  pgtype.Timestamptz
	ReadAt       pgtype.Timestamptz
}

type SchemaMigration struct {
	Version string
}

type User struct {
	ID        pgtype.UUID
	Name      string
	PublicKey []byte
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}
