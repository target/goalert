-- +migrate Up
CREATE OR REPLACE FUNCTION fn_job_signal ()
    RETURNS TRIGGER
    AS $$
DECLARE
    key_name text := 'signal_job_' || replace(NEW.service_id::text, '-', '_');
BEGIN
    -- Only run once per transaction per ID
    IF current_setting('local.' || key_name, TRUE) IS NOT NULL THEN
        RETURN NULL;
    END IF;
    -- Mark this ID as processed in this transaction
    PERFORM
        set_config('local.' || key_name, 'true', TRUE);
    INSERT INTO river_job (args, kind, max_attempts, priority)
        VALUES (json_build_object('ServiceID', NEW.service_id), 'signal-manager-schedule-outgoing-messages', 25, 2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', '{"queue":"engine-signal-mgr"}');
    RETURN NULL;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_pending_signals_after_insert
    AFTER INSERT ON pending_signals
    FOR EACH ROW
    EXECUTE FUNCTION fn_job_signal ();

-- +migrate Down
DROP TRIGGER IF EXISTS trg_pending_signals_after_insert ON pending_signals;

DROP FUNCTION IF EXISTS fn_job_signal ();

