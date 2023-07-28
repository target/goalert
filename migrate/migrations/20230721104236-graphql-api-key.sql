-- +migrate Up
CREATE TABLE api_keys(
    id uuid PRIMARY KEY,
    name text NOT NULL UNIQUE,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    service_id uuid REFERENCES services(id) ON DELETE CASCADE,
    policy jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    expires_at timestamp with time zone NOT NULL,
    last_used_at timestamp with time zone
);

-- +migrate Down
DROP TABLE api_keys;

