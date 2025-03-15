-- name: InsertMessage :one
INSERT INTO messages (sender, recipient, cipher_sym_key, ciphertext, sent_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetUndeliveredMessages :many
SELECT * FROM messages
WHERE
	recipient = ? AND
	delivered_at IS NULL;

-- name: SetMessageSent :exec
UPDATE messages SET sent_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: SetMessageDelivered :exec
UPDATE messages SET delivered_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: SetMessageRead :exec
UPDATE messages SET read_at = CURRENT_TIMESTAMP WHERE id = ?;
