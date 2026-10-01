package store

// schema contains the SQLite DDL for Open-Skill.
//
// Identity model: user owns many players (handles).
// Ratings are materialized per (community_id, player_id, mode) as a projection
// of the match log. Mode is derived from the participant set at report time.
const schema = `
CREATE TABLE IF NOT EXISTS users (
	id          TEXT PRIMARY KEY,
	display     TEXT NOT NULL,
	created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS players (
	id          TEXT PRIMARY KEY,
	user_id     TEXT NOT NULL REFERENCES users(id),
	handle      TEXT NOT NULL UNIQUE,
	created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS communities (
	id          TEXT PRIMARY KEY,
	name        TEXT NOT NULL,
	mode        TEXT NOT NULL DEFAULT 'open',   -- 'open' | 'approval'
	draws       INTEGER NOT NULL DEFAULT 1,
	code        TEXT NOT NULL UNIQUE,
	owner_id    TEXT NOT NULL REFERENCES users(id),
	created_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS community_members (
	community_id TEXT NOT NULL REFERENCES communities(id),
	user_id      TEXT NOT NULL REFERENCES users(id),
	role         TEXT NOT NULL DEFAULT 'member', -- 'owner'|'admin'|'member'
	joined_at    TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (community_id, user_id)
);

-- Materialized skill rating per community, player, and mode (projection of the
-- match log). OpenSkill state is represented by mu and sigma.
CREATE TABLE IF NOT EXISTS community_join_requests (
	community_id TEXT NOT NULL REFERENCES communities(id),
	user_id TEXT NOT NULL REFERENCES users(id),
	status TEXT NOT NULL DEFAULT 'pending',
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (community_id, user_id)
);

CREATE TABLE IF NOT EXISTS community_integrations (
	community_id TEXT PRIMARY KEY REFERENCES communities(id),
	discord TEXT NOT NULL DEFAULT '',
	webhook TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS ratings (
	community_id  TEXT NOT NULL REFERENCES communities(id),
	player_id     TEXT NOT NULL REFERENCES players(id),
	mode          TEXT NOT NULL CHECK (mode IN ('1v1', 'team', 'ffa')),
	mu            REAL NOT NULL DEFAULT 25,
	sigma         REAL NOT NULL DEFAULT 8.3333333333,
	games         INTEGER NOT NULL DEFAULT 0,
	wins          INTEGER NOT NULL DEFAULT 0,
	losses        INTEGER NOT NULL DEFAULT 0,
	draws         INTEGER NOT NULL DEFAULT 0,
	last_played   TEXT,
	PRIMARY KEY (community_id, player_id, mode)
);

CREATE TABLE IF NOT EXISTS matches (
	id          TEXT PRIMARY KEY,
	community_id TEXT NOT NULL REFERENCES communities(id),
	mode        TEXT NOT NULL CHECK (mode IN ('1v1', 'team', 'ffa')),
	played_at   TEXT NOT NULL DEFAULT (datetime('now')),
	notes       TEXT
);

CREATE TABLE IF NOT EXISTS match_participants (
	match_id       TEXT NOT NULL REFERENCES matches(id),
	player_id      TEXT NOT NULL REFERENCES players(id),
	team           TEXT,                       -- team label for team/ffa grouping
	score          REAL DEFAULT 0,             -- points scored
	finish         INTEGER,                    -- 0-based finish for ffa
	rating_delta   REAL NOT NULL DEFAULT 0,    -- display-rating delta applied
	PRIMARY KEY (match_id, player_id)
);

CREATE INDEX IF NOT EXISTS idx_ratings_community_mode_mu ON ratings(community_id, mode, mu DESC);
CREATE INDEX IF NOT EXISTS idx_match_participants_match ON match_participants(match_id);
CREATE INDEX IF NOT EXISTS idx_match_participants_player ON match_participants(player_id);
`
