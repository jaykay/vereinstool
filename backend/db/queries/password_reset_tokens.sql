-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (token, user_id, expires_at)
VALUES (?, ?, ?);

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE token = ? AND used = 0 AND expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens SET used = 1 WHERE token = ?;

-- name: DeleteExpiredPasswordResetTokens :exec
DELETE FROM password_reset_tokens WHERE expires_at < CURRENT_TIMESTAMP OR used = 1;
