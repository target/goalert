
-- +migrate Up

CREATE TABLE auth_token_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users (id),
    expires_at TIMESTAMP NOT NULL DEFAULT now()+'5 minutes'::INTERVAL,
    user_agent TEXT NOT NULL
);

-- +migrate Down

DROP TABLE auth_token_codes;
