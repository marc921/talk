-- name: ListLocalUsers :many
SELECT * FROM local_users;

-- name: GetLocalUserByName :one
SELECT * FROM local_users WHERE name = ?;

-- name: InsertLocalUser :one
INSERT INTO local_users (name, private_key) VALUES (?, ?) RETURNING *;