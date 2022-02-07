-- +migrate Up

CREATE TABLE alert_metrics (
    alert_id INT PRIMARY KEY REFERENCES alerts (id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    time_to_ack INTERVAL,
    time_to_close INTERVAL NOT NULL,
    escalated BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE INDEX idx_closed_events ON alert_logs (timestamp) WHERE event = 'closed';

ALTER TYPE engine_processing_type ADD VALUE IF NOT EXISTS 'metrics';
INSERT INTO engine_processing_versions (type_id) VALUES ('metrics');

ALTER TABLE engine_processing_versions ADD COLUMN state JSONB NOT NULL DEFAULT '{}'::jsonb;


-- +migrate Down
ALTER TABLE engine_processing_versions DROP COLUMN state;
DELETE FROM engine_processing_versions WHERE type_id = 'metrics';
DROP INDEX idx_closed_events;
DROP TABLE alert_metrics;
