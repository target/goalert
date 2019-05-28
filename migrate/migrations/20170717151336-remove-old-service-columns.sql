
-- +migrate Up
ALTER TABLE service
    DROP COLUMN created_at,
    DROP COLUMN summary,
    DROP COLUMN type,
    DROP COLUMN self,
    DROP COLUMN html_url,
    DROP COLUMN status,
    DROP COLUMN last_incident_timestamp,
    DROP COLUMN conference_url,
    DROP COLUMN dialin_number,
    DROP COLUMN acknowledgement_timeout,
    DROP COLUMN auto_resolve_timeout,
    DROP COLUMN maintenance_mode,
    DROP COLUMN incident_urgency_type,
    DROP COLUMN incident_urgency_value;

-- +migrate Down

ALTER TABLE service
    ADD COLUMN created_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN summary TEXT,
    ADD COLUMN type TEXT,
    ADD COLUMN self TEXT,
    ADD COLUMN html_url TEXT,
    ADD COLUMN status TEXT,
    ADD COLUMN last_incident_timestamp TIMESTAMP WITH TIME ZONE,
    ADD COLUMN conference_url TEXT,
    ADD COLUMN dialin_number TEXT,
    ADD COLUMN acknowledgement_timeout INT,
    ADD COLUMN auto_resolve_timeout INT,
    ADD COLUMN maintenance_mode BOOLEAN,
    ADD COLUMN incident_urgency_type TEXT,
    ADD COLUMN incident_urgency_value TEXT;
