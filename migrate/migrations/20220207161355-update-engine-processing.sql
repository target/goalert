-- +migrate Up

INSERT INTO engine_processing_versions (type_id) VALUES ('metrics');

ALTER TABLE engine_processing_versions ADD COLUMN state JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +migrate Down

ALTER TABLE engine_processing_versions DROP COLUMN state;
DELETE FROM engine_processing_versions WHERE type_id = 'metrics';