-- +migrate Up
CREATE OR REPLACE FUNCTION fn_track_alert_status ()
    RETURNS TRIGGER
    AS $$
BEGIN
    INSERT INTO river_job (args, kind, max_attempts, priority)
        VALUES (json_build_object('AlertID', NEW.id), 'status-manager-look-for-work', 25, 2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', '{"queue":"status-manager"}');
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_track_alert_status_update
    AFTER UPDATE ON alerts
    FOR EACH ROW
    WHEN (NEW.status IS DISTINCT FROM OLD.status)
    EXECUTE FUNCTION fn_track_alert_status ();

-- +migrate Down
DROP TRIGGER IF EXISTS trg_track_alert_status_update ON alerts;

DROP FUNCTION IF EXISTS fn_track_alert_status ();

