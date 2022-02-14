-- +migrate Up

CREATE TABLE alert_metrics (
    id BIGINT PRIMARY KEY REFERENCES alerts (id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    time_to_ack INTERVAL,
    time_to_close INTERVAL,
    escalated BOOLEAN DEFAULT FALSE NOT NULL
);

CREATE INDEX idx_closed_events ON alert_logs (timestamp) WHERE event = 'closed';
ALTER TABLE engine_processing_versions ADD COLUMN state JSONB NOT NULL DEFAULT '{}'::jsonb;
INSERT INTO engine_processing_versions (type_id, version) VALUES ('metrics', 1);

-- +migrate Down

ALTER TABLE engine_processing_versions DROP COLUMN state;
DELETE FROM engine_processing_versions WHERE type_id = 'metrics';
DROP TABLE alert_metrics;
