-- name: InsertUser :one
INSERT INTO users (name, public_key) 
VALUES (?, ?)
ON CONFLICT(name) DO UPDATE SET
    updated_at = updated_at
RETURNING *, 
    changes() = 1;

-- name: GetUser :one
SELECT * FROM users WHERE name = ?;

-- name: ListUsers :many
SELECT * FROM users;