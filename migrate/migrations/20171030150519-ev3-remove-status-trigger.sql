
-- +migrate Up

DROP TRIGGER log_alert_status_changed ON alerts;
DROP FUNCTION log_alert_status_changed_insert();

-- +migrate Down

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION log_alert_status_changed_insert() RETURNS TRIGGER AS
    $$
        BEGIN
            IF NEW.status = 'closed'::enum_alert_status THEN
                INSERT INTO alert_logs (alert_id, event, message) VALUES (
                    NEW.id, 'closed'::enum_alert_log_event, 'Closed'
                );
            ELSIF OLD.status = 'closed'::enum_alert_status THEN
                INSERT INTO alert_logs (alert_id, event, message) VALUES (
                    NEW.id, 'reopened'::enum_alert_log_event, 'Reopened as '||NEW.status::TEXT
                );
            ELSE
                INSERT INTO alert_logs (alert_id, event, message) VALUES (
                    NEW.id, 'status_changed'::enum_alert_log_event, 'Status updated from '||OLD.status::TEXT||' to '||NEW.status::TEXT
                );
            END IF;
            RETURN NEW;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


CREATE TRIGGER log_alert_status_changed
    AFTER UPDATE ON alerts
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status)
    EXECUTE PROCEDURE log_alert_status_changed_insert();
