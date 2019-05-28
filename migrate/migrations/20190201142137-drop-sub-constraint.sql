
-- +migrate Up
ALTER TABLE alert_logs
    DROP CONSTRAINT alert_logs_one_subject;

DROP TRIGGER trg_insert_alert_logs_user_last_alert ON alert_logs;
CREATE TRIGGER trg_insert_alert_logs_user_last_alert
AFTER INSERT
ON alert_logs
FOR EACH ROW
WHEN (NEW.event = 'notification_sent' AND NEW.sub_type = 'user')
EXECUTE PROCEDURE fn_insert_user_last_alert_log();


-- +migrate Down
ALTER TABLE alert_logs
    ADD CONSTRAINT alert_logs_one_subject CHECK (NOT (sub_user_id IS NOT NULL AND sub_integration_key_id IS NOT NULL AND sub_hb_monitor_id IS NOT NULL));

DROP TRIGGER trg_insert_alert_logs_user_last_alert ON alert_logs;
CREATE TRIGGER trg_insert_alert_logs_user_last_alert
AFTER INSERT
ON alert_logs
FOR EACH ROW
WHEN (NEW.event = 'notification_sent')
EXECUTE PROCEDURE fn_insert_user_last_alert_log();
