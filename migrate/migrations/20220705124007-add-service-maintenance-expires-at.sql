-- +migrate Up
UPDATE engine_processing_versions SET "version" = 4 WHERE type_id = 'escalation';
ALTER TABLE services ADD COLUMN maintenance_expires_at TIMESTAMPTZ;

-- +migrate Down
UPDATE engine_processing_versions SET "version" = 3 WHERE type_id = 'escalation';
ALTER TABLE services DROP COLUMN maintenance_expires_at;
