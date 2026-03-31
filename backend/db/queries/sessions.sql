-- name: GetSession :one
SELECT * FROM sessions WHERE id = ? AND expires_at > CURRENT_TIMESTAMP LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP;

-- name: DeleteUserSessions :exec
DELETE FROM sessions WHERE user_id = ?;
