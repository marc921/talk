// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: local_users.sql

package sqlcgen

import (
	"context"
)

const insertLocalUser = `-- name: InsertLocalUser :one
INSERT INTO local_users (name, private_key) VALUES (?, ?) RETURNING name, private_key
`

type InsertLocalUserParams struct {
	Name       string
	PrivateKey []byte
}

func (q *Queries) InsertLocalUser(ctx context.Context, arg InsertLocalUserParams) (LocalUser, error) {
	row := q.db.QueryRowContext(ctx, insertLocalUser, arg.Name, arg.PrivateKey)
	var i LocalUser
	err := row.Scan(&i.Name, &i.PrivateKey)
	return i, err
}

const listLocalUsers = `-- name: ListLocalUsers :many
SELECT name, private_key FROM local_users
`

func (q *Queries) ListLocalUsers(ctx context.Context) ([]LocalUser, error) {
	rows, err := q.db.QueryContext(ctx, listLocalUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LocalUser
	for rows.Next() {
		var i LocalUser
		if err := rows.Scan(&i.Name, &i.PrivateKey); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}