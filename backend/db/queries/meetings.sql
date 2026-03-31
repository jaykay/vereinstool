-- name: GetMeeting :one
SELECT * FROM meetings WHERE id = ? LIMIT 1;

-- name: ListMeetings :many
SELECT * FROM meetings ORDER BY scheduled_at DESC;

-- name: ListMeetingsByStatus :many
SELECT * FROM meetings WHERE status = ? ORDER BY scheduled_at DESC;

-- name: CreateMeeting :one
INSERT INTO meetings (title, scheduled_at, duration_mins, location, created_by)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateMeeting :exec
UPDATE meetings SET title = ?, scheduled_at = ?, duration_mins = ?, location = ? WHERE id = ?;

-- name: UpdateMeetingStatus :exec
UPDATE meetings SET status = ? WHERE id = ?;
