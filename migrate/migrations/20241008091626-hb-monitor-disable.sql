-- +migrate Up
ALTER TABLE heartbeat_monitors
    ADD COLUMN muted text;

UPDATE
    engine_processing_versions
SET
    version = 2
WHERE
    type_id = 'heartbeat';

-- +migrate Down
ALTER TABLE heartbeat_monitors
    DROP COLUMN muted;

UPDATE
    engine_processing_versions
SET
    version = 1
WHERE
    type_id = 'heartbeat';

