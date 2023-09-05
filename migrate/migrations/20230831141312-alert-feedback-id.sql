-- +migrate Up
ALTER TABLE alert_feedback
    ADD COLUMN id BIGSERIAL UNIQUE NOT NULL;

-- +migrate Down
ALTER TABLE alert_feedback
    DROP COLUMN id;

