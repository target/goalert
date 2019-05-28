
-- +migrate Up

CREATE TABLE keyring (
    id TEXT PRIMARY KEY,
    verification_keys BYTEA NOT NULL,
    signing_key BYTEA NOT NULL,
    next_key BYTEA NOT NULL,
    next_rotation TIMESTAMP WITH TIME ZONE,
    rotation_count BIGINT NOT NULL
);

-- +migrate Down

DROP TABLE keyring;
