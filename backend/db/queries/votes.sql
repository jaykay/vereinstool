-- name: GetVote :one
SELECT * FROM votes WHERE topic_id = ? AND user_id = ? LIMIT 1;

-- name: CreateVote :exec
INSERT INTO votes (topic_id, user_id) VALUES (?, ?);

-- name: DeleteVote :exec
DELETE FROM votes WHERE topic_id = ? AND user_id = ?;

-- name: ListVotesByTopic :many
SELECT * FROM votes WHERE topic_id = ?;

-- name: ListUserVotesForMeeting :many
SELECT v.topic_id FROM votes v
JOIN topics t ON t.id = v.topic_id
WHERE t.meeting_id = ? AND v.user_id = ?;

-- name: ListUserVotesForUnassigned :many
SELECT v.topic_id FROM votes v
JOIN topics t ON t.id = v.topic_id
WHERE t.meeting_id IS NULL AND v.user_id = ?;
