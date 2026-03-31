-- name: GetTopic :one
SELECT * FROM topics WHERE id = ? LIMIT 1;

-- name: ListTopicsByMeeting :many
SELECT * FROM topics WHERE meeting_id = ? ORDER BY vote_count DESC, created_at ASC;

-- name: CreateTopic :one
INSERT INTO topics (meeting_id, title, description, category, submitted_by, estimated_mins)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateTopic :exec
UPDATE topics SET title = ?, description = ?, category = ?, estimated_mins = ?, status = ? WHERE id = ?;

-- name: DeleteTopic :exec
DELETE FROM topics WHERE id = ?;

-- name: IncrementVoteCount :exec
UPDATE topics SET vote_count = vote_count + 1 WHERE id = ?;

-- name: DecrementVoteCount :exec
UPDATE topics SET vote_count = vote_count - 1 WHERE id = ?;
