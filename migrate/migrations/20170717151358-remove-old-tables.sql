
-- +migrate Up
DROP TABLE team_user, team, integration;

-- +migrate Down

CREATE TABLE team (
    id TEXT PRIMARY KEY,
    name TEXT,
    description TEXT
);

CREATE TABLE team_user (
    id TEXT PRIMARY KEY,
    team_id TEXT REFERENCES team (id),
    user_id UUID REFERENCES users (id)
);

CREATE TABLE integration (
    id TEXT PRIMARY KEY,
    type TEXT,
    name TEXT,
    integration_key TEXT UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE,
    service_id TEXT REFERENCES service (id) ON DELETE CASCADE
);

