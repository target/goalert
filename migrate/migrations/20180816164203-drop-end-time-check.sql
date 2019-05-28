
-- +migrate Up
ALTER TABLE user_overrides DROP CONSTRAINT IF EXISTS user_overrides_end_time_check;

-- +migrate Down
ALTER TABLE user_overrides
    ADD CONSTRAINT user_overrides_end_time_check CHECK ((end_time > now()));
