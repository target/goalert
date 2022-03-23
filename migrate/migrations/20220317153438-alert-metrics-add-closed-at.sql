-- +migrate Up

UPDATE engine_processing_versions
SET "version" = 2, state = DEFAULT
WHERE type_id = 'metrics';

ALTER TABLE alert_metrics ADD COLUMN closed_at TIMESTAMP WITH TIME ZONE NOT NULL;

TRUNCATE alert_metrics;


-- +migrate Down

ALTER TABLE alert_metrics DROP COLUMN closed_at;

UPDATE engine_processing_versions
SET "version" = 1, state = DEFAULT
WHERE type_id = 'metrics';

TRUNCATE alert_metrics;
