-- name: AddAttendee :exec
INSERT OR IGNORE INTO meeting_attendees (meeting_id, user_id) VALUES (?, ?);

-- name: RemoveAttendee :exec
DELETE FROM meeting_attendees WHERE meeting_id = ? AND user_id = ?;

-- name: ListAttendees :many
SELECT u.id, u.email, u.name, u.role, ma.present
FROM meeting_attendees ma
JOIN users u ON u.id = ma.user_id
WHERE ma.meeting_id = ?
ORDER BY u.name;

-- name: SetAttendeePresent :exec
UPDATE meeting_attendees SET present = ? WHERE meeting_id = ? AND user_id = ?;
