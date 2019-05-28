
-- +migrate Up

CREATE TYPE enum_alert_log_subject_type AS ENUM (
    'user',
    'integration_key'
);

ALTER TABLE alert_logs
    ADD COLUMN sub_type enum_alert_log_subject_type,
    ADD COLUMN sub_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN sub_integration_key_id UUID REFERENCES integration_keys (id) ON DELETE SET NULL,
    ADD COLUMN sub_classifier TEXT NOT NULL DEFAULT '',
    ADD COLUMN meta JSON,
    ADD CONSTRAINT alert_logs_one_subject CHECK(
        NOT (sub_user_id IS NOT NULL AND sub_integration_key_id IS NOT NULL)
    )
;

ALTER TABLE alerts
    ADD COLUMN created_at TIMESTAMP WITH TIME ZONE;

UPDATE alerts alert
SET created_at = "timestamp"
FROM alert_logs log
WHERE
    log.alert_id = alert.id AND
    log."event" = 'created';

ALTER TABLE alerts
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN created_at SET DEFAULT now();

DROP TRIGGER log_alert_creation ON alerts;
DROP FUNCTION log_alert_creation_insert();

-- +migrate Down

ALTER TABLE alert_logs
    DROP COLUMN sub_type,
    DROP COLUMN sub_user_id,
    DROP COLUMN sub_integration_key_id,
    DROP COLUMN sub_classifier,
    DROP COLUMN meta;

ALTER TABLE alerts
    DROP COLUMN created_at;

DROP TYPE enum_alert_log_subject_type;

-- +migrate StatementBegin
CREATE FUNCTION log_alert_creation_insert() RETURNS TRIGGER AS
    $$
        BEGIN
            INSERT INTO alert_logs (alert_id, event, message) VALUES (
                NEW.id, 'created'::enum_alert_log_event, 'Created via: '||NEW.source::TEXT
            );
            RETURN NEW;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER log_alert_creation
    AFTER INSERT ON alerts
    FOR EACH ROW
    EXECUTE PROCEDURE log_alert_creation_insert();