-- +migrate Up
CREATE TABLE gql_api_keys(
    id uuid PRIMARY KEY,
    name text NOT NULL UNIQUE,
    description text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    created_by uuid REFERENCES users(id) ON DELETE SET NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
    -- We must use json instead of jsonb because we need to be able to compute a reproducable hash of the policy
    -- jsonb will not work because it does not guarantee a stable order of keys or whitespace consistency.
    --
    -- We also don't need to be able to query the policy, so json is fine.
    policy json NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by uuid REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE gql_api_key_usage(
    id bigserial PRIMARY KEY,
    api_key_id uuid REFERENCES gql_api_keys(id) ON DELETE CASCADE UNIQUE,
    used_at timestamp with time zone NOT NULL DEFAULT now(),
    user_agent text,
    ip_address inet
);

-- +migrate Down
DROP TABLE gql_api_key_usage;

DROP TABLE gql_api_keys;

