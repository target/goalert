-- +migrate Up
UPDATE engine_processing_versions SET version = 2 WHERE type_id = 'status_update';

-- don't update existing alerts
DROP TRIGGER trg_insert_alert_logs_user_last_alert_update ON alert_logs;
-- don't create any new logs
DROP TRIGGER trg_insert_alert_logs_user_last_alert ON alert_logs;

-- +migrate Down
UPDATE engine_processing_versions SET version = 1 WHERE type_id = 'status_update';

CREATE TRIGGER trg_insert_alert_logs_user_last_alert
AFTER INSERT
ON alert_logs
FOR EACH ROW
WHEN (NEW.event = 'notification_sent')
EXECUTE PROCEDURE fn_insert_user_last_alert_log();

CREATE TRIGGER trg_insert_alert_logs_user_last_alert_update
AFTER INSERT
ON alert_logs
FOR EACH ROW
WHEN (NEW.event IN ('acknowledged', 'closed'))
EXECUTE PROCEDURE fn_update_user_last_alert_log();