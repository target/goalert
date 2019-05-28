
-- +migrate Up

ALTER TABLE schedules
    ADD COLUMN last_processed TIMESTAMP WITH TIME ZONE;

ALTER TABLE schedule_rules
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE process_schedules (
    schedule_id UUID PRIMARY KEY REFERENCES schedules (id) ON DELETE CASCADE,
    client_id UUID,
    deadline TIMESTAMP WITH TIME ZONE,
    last_processed TIMESTAMP WITH TIME ZONE
);

CREATE INDEX process_schedules_oldest_first ON process_schedules (last_processed ASC NULLS FIRST);

-- +migrate Down

DROP TABLE process_schedules;

ALTER TABLE schedules
    DROP COLUMN last_processed;

ALTER TABLE schedule_rules
    DROP COLUMN is_active;
