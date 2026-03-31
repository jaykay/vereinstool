CREATE TABLE users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    email         TEXT NOT NULL UNIQUE,
    name          TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'member',
    is_active     INTEGER NOT NULL DEFAULT 1,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions (
    id         TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE password_reset_tokens (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL,
    used       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE meetings (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    title        TEXT NOT NULL,
    scheduled_at DATETIME NOT NULL,
    location     TEXT,
    status       TEXT NOT NULL DEFAULT 'open',
    created_by   INTEGER NOT NULL REFERENCES users(id),
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE meeting_attendees (
    meeting_id INTEGER NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    present    INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (meeting_id, user_id)
);

CREATE TABLE topics (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    meeting_id     INTEGER REFERENCES meetings(id) ON DELETE SET NULL,
    title          TEXT NOT NULL,
    description    TEXT,
    category       TEXT,
    submitted_by   INTEGER NOT NULL REFERENCES users(id),
    estimated_mins INTEGER NOT NULL DEFAULT 10,
    status         TEXT NOT NULL DEFAULT 'open',
    vote_count     INTEGER NOT NULL DEFAULT 0,
    position       INTEGER,
    is_recurring   INTEGER NOT NULL DEFAULT 0,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE votes (
    topic_id   INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (topic_id, user_id)
);

CREATE TABLE decisions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id      INTEGER NOT NULL REFERENCES topics(id),
    meeting_id    INTEGER NOT NULL REFERENCES meetings(id),
    text          TEXT NOT NULL,
    votes_yes     INTEGER,
    votes_no      INTEGER,
    votes_abstain INTEGER,
    recorded_by   INTEGER NOT NULL REFERENCES users(id),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tasks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    topic_id    INTEGER REFERENCES topics(id) ON DELETE SET NULL,
    meeting_id  INTEGER REFERENCES meetings(id) ON DELETE SET NULL,
    title       TEXT NOT NULL,
    description TEXT,
    assigned_to INTEGER REFERENCES users(id) ON DELETE SET NULL,
    due_date    DATE,
    status      TEXT NOT NULL DEFAULT 'open',
    created_by  INTEGER NOT NULL REFERENCES users(id),
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_topics_meeting_id ON topics(meeting_id);
CREATE INDEX idx_topics_vote_count ON topics(vote_count DESC);
CREATE INDEX idx_decisions_meeting_id ON decisions(meeting_id);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
