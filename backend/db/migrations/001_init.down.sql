DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_tasks_assigned_to;
DROP INDEX IF EXISTS idx_decisions_meeting_id;
DROP INDEX IF EXISTS idx_topics_vote_count;
DROP INDEX IF EXISTS idx_topics_meeting_id;

DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS decisions;
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS topics;
DROP TABLE IF EXISTS meeting_attendees;
DROP TABLE IF EXISTS meetings;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
