-- name: ListPublicUsers :many
SELECT * FROM public_users;

-- name: GetPublicUserByName :one
SELECT * FROM public_users WHERE name = ?;

-- name: InsertPublicUser :one
INSERT INTO public_users (name, public_key) VALUES (?, ?) RETURNING *;