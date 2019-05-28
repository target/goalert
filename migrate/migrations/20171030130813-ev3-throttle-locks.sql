
-- +migrate Up

CREATE TABLE process_alerts (
    alert_id BIGINT PRIMARY KEY REFERENCES alerts(id) ON DELETE CASCADE,
    client_id UUID,
    deadline TIMESTAMP WITH TIME ZONE,
    last_processed TIMESTAMP WITH TIME ZONE
);

CREATE INDEX process_alerts_oldest_first ON process_alerts (last_processed ASC NULLS FIRST);

CREATE TABLE process_rotations (
    rotation_id UUID PRIMARY KEY REFERENCES rotations(id) ON DELETE CASCADE,
    client_id UUID,
    deadline TIMESTAMP WITH TIME ZONE,
    last_processed TIMESTAMP WITH TIME ZONE
);

CREATE INDEX process_rotations_oldest_first ON process_rotations (last_processed ASC NULLS FIRST);

ALTER TABLE alerts
    ADD COLUMN last_processed TIMESTAMP WITH TIME ZONE;

ALTER TABLE rotations
    ADD COLUMN last_processed TIMESTAMP WITH TIME ZONE;

UPDATE alerts
SET last_processed = last_action_time
FROM throttle
WHERE status != 'closed';

-- +migrate Down

DROP TABLE process_alerts;
DROP TABLE process_rotations;

ALTER TABLE alerts
    DROP COLUMN last_processed;

ALTER TABLE rotations
    DROP COLUMN last_processed;
