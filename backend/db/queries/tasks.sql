-- name: GetTask :one
SELECT * FROM tasks WHERE id = ? LIMIT 1;

-- name: ListTasks :many
SELECT * FROM tasks ORDER BY created_at DESC;

-- name: ListTasksByAssignee :many
SELECT * FROM tasks WHERE assigned_to = ? ORDER BY due_date ASC;

-- name: ListTasksByStatus :many
SELECT * FROM tasks WHERE status = ? ORDER BY due_date ASC;

-- name: CreateTask :one
INSERT INTO tasks (topic_id, meeting_id, title, description, assigned_to, due_date, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateTask :exec
UPDATE tasks SET title = ?, description = ?, assigned_to = ?, due_date = ?, status = ? WHERE id = ?;
