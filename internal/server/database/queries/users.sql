-- name: InsertUser :one
INSERT INTO users (name, public_key) 
VALUES ($1, $2)
ON CONFLICT(name) DO NOTHING
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE name = $1;

-- name: ListUsers :many
SELECT * FROM users;