
-- +migrate Up

CREATE TABLE auth_nonce (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- +migrate Down

DROP TABLE auth_nonce;
