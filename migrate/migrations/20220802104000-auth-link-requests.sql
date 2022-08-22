-- +migrate Up
CREATE TABLE auth_link_requests (
    id UUID PRIMARY KEY,
    provider_id TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB
);

-- +migrate Down
DROP TABLE auth_link_requests;
