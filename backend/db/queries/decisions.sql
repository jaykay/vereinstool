-- name: GetDecision :one
SELECT * FROM decisions WHERE id = ? LIMIT 1;

-- name: ListDecisions :many
SELECT * FROM decisions ORDER BY created_at DESC;

-- name: ListDecisionsByMeeting :many
SELECT * FROM decisions WHERE meeting_id = ? ORDER BY created_at;

-- name: CreateDecision :one
INSERT INTO decisions (topic_id, meeting_id, text, votes_yes, votes_no, votes_abstain, recorded_by)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateDecision :exec
UPDATE decisions SET text = ?, votes_yes = ?, votes_no = ?, votes_abstain = ? WHERE id = ?;
