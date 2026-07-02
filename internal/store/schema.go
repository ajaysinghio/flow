package store

const schema = `
CREATE TABLE IF NOT EXISTS migrations (
	version    INTEGER PRIMARY KEY,
	applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tasks (
	id           TEXT    PRIMARY KEY,
	title        TEXT    NOT NULL,
	size         TEXT    NOT NULL DEFAULT 'm',
	energy       TEXT    NOT NULL DEFAULT 'med',
	status       TEXT    NOT NULL DEFAULT 'todo',
	parent_id    TEXT    REFERENCES tasks(id),
	tags         TEXT    NOT NULL DEFAULT '[]',
	created_at   DATETIME NOT NULL DEFAULT (datetime('now')),
	completed_at DATETIME
);

CREATE TABLE IF NOT EXISTS checkins (
	id        TEXT    PRIMARY KEY,
	mood      INTEGER NOT NULL,
	energy    INTEGER NOT NULL,
	note      TEXT    NOT NULL DEFAULT '',
	timestamp DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS notes (
	id        TEXT    PRIMARY KEY,
	content   TEXT    NOT NULL,
	tags      TEXT    NOT NULL DEFAULT '[]',
	timestamp DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS focus_sessions (
	id          TEXT    PRIMARY KEY,
	task_id     TEXT    REFERENCES tasks(id),
	started_at  DATETIME NOT NULL DEFAULT (datetime('now')),
	ended_at    DATETIME,
	interrupted INTEGER NOT NULL DEFAULT 0
);
`
