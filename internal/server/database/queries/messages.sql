-- name: InsertMessage :one
INSERT INTO messages (sender, recipient, cipher_sym_key, ciphertext, sent_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetUndeliveredMessages :many
SELECT * FROM messages
WHERE
	recipient = $1 AND
	delivered_at IS NULL;

-- name: SetMessageSent :exec
UPDATE messages SET sent_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: SetMessageDelivered :exec
UPDATE messages SET delivered_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: SetMessageRead :exec
UPDATE messages SET read_at = CURRENT_TIMESTAMP WHERE id = $1;
