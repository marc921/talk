-- name: ListConversations :many
SELECT * FROM conversations WHERE local_user_name = ?;

-- name: GetConversation :one
SELECT * FROM conversations WHERE local_user_name = ? AND remote_user_name = ?;

-- name: InsertConversation :one
INSERT INTO conversations (local_user_name, remote_user_name) 
VALUES (?, ?) 
ON CONFLICT (local_user_name, remote_user_name) 
DO UPDATE SET local_user_name = EXCLUDED.local_user_name
RETURNING *;