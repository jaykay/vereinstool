-- name: GetVote :one
SELECT * FROM votes WHERE topic_id = ? AND user_id = ? LIMIT 1;

-- name: CreateVote :exec
INSERT INTO votes (topic_id, user_id) VALUES (?, ?);

-- name: DeleteVote :exec
DELETE FROM votes WHERE topic_id = ? AND user_id = ?;

-- name: ListVotesByTopic :many
SELECT * FROM votes WHERE topic_id = ?;
