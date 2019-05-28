
-- +migrate Up
DROP TABLE auth_token_codes;

-- +migrate Down
CREATE TABLE auth_token_codes (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id uuid NOT NULL UNIQUE REFERENCES users(id),
    expires_at timestamp without time zone NOT NULL DEFAULT (now() + '00:05:00'::interval),
    user_agent text NOT NULL
);
