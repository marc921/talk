-- name: ListMessages :many
SELECT * FROM messages WHERE conversation_id = ?;

-- name: InsertMessage :one
INSERT INTO messages (
	conversation_id,
	sender,
	receiver,
	content
) VALUES (?, ?, ?, ?) RETURNING *;

-- name: MarkMessageSent :one
UPDATE messages SET sent_at = ? WHERE id = ? RETURNING *;

-- name: MarkMessageDelivered :one
UPDATE messages SET delivered_at = ? WHERE id = ? RETURNING *;

-- name: MarkMessageRead :one
UPDATE messages SET read_at = ? WHERE id = ? RETURNING *;