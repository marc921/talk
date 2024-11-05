-- name: ListLocalUsers :many
SELECT * FROM local_users;

-- name: InsertLocalUser :one
INSERT INTO local_users (name, private_key) VALUES (?, ?) RETURNING *;