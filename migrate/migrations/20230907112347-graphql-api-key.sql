-- +migrate Up
CREATE TABLE gql_api_keys(
    id uuid PRIMARY KEY,
    name text NOT NULL UNIQUE,
    description text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    created_by uuid REFERENCES users(id) ON DELETE SET NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
    policy jsonb NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by uuid REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE gql_api_key_usage(
    id bigserial PRIMARY KEY,
    api_key_id uuid REFERENCES gql_api_keys(id) ON DELETE CASCADE,
    used_at timestamp with time zone NOT NULL DEFAULT now(),
    user_agent text,
    ip_address inet
);

CREATE INDEX idx_gql_most_recent_use ON gql_api_key_usage(api_key_id, used_at DESC);

-- +migrate Down
DROP TABLE gql_api_key_usage;

DROP TABLE gql_api_keys;

