-- +migrate Up
CREATE TABLE uik_config(
    id uuid PRIMARY KEY REFERENCES integration_keys(id) ON DELETE CASCADE,
    config jsonb NOT NULL
);

-- +migrate Down
DROP TABLE uik_config;

