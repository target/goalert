-- +migrate Up
ALTER TABLE heartbeat_monitors
    ADD COLUMN additional_details TEXT;

-- +migrate Down
ALTER TABLE heartbeat_monitors
    DROP COLUMN additional_details;

