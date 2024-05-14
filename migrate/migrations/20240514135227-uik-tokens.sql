-- +migrate Up
ALTER TABLE uik_config
    ADD COLUMN primary_token uuid UNIQUE,
    ADD COLUMN primary_token_hint text,
    ADD COLUMN secondary_token uuid UNIQUE,
    ADD COLUMN secondary_token_hint text;

-- +migrate Down
ALTER TABLE uik_config
    DROP COLUMN primary_token, primary_token_hint, secondary_token, secondary_token_hint;

