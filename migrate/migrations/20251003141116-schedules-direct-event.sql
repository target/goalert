-- +migrate Up
CREATE OR REPLACE FUNCTION fn_job_schedule(id uuid)
    RETURNS VOID
    AS $$
DECLARE
    key_name text := 'schedule_job_' || id::text;
BEGIN
    -- Only run once per transaction per ID
    IF current_setting('local.' || key_name, TRUE) IS NOT NULL THEN
        RETURN;
    END IF;
    -- Mark this ID as processed in this transaction
    PERFORM
        set_config('local.' || key_name, 'true', TRUE);
    INSERT INTO river_job(args, kind, max_attempts, priority)
        VALUES (json_build_object('ScheduleID', id), 'schedule-manager-update', 25, 2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', '{"queue":"schedule-manager"}');
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_track_schedule_updates()
    RETURNS TRIGGER
    AS $$
BEGIN
    IF TG_TABLE_NAME = 'schedules' THEN
        PERFORM
            fn_job_schedule(NEW.id);
    ELSIF TG_OP = 'DELETE' THEN
        PERFORM
            fn_job_schedule(OLD.schedule_id);
        RETURN OLD;
    ELSE
        PERFORM
            fn_job_schedule(NEW.schedule_id);
    END IF;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_track_schedule_updates
    AFTER INSERT OR UPDATE ON schedules
    FOR EACH ROW
    EXECUTE FUNCTION fn_track_schedule_updates();

CREATE TRIGGER trg_track_schedule_data_updates
    AFTER INSERT OR UPDATE OR DELETE ON schedule_data
    FOR EACH ROW
    EXECUTE FUNCTION fn_track_schedule_updates();

CREATE TRIGGER trg_track_schedule_rules_updates
    AFTER INSERT OR UPDATE OR DELETE ON schedule_rules
    FOR EACH ROW
    EXECUTE FUNCTION fn_track_schedule_updates();

CREATE TRIGGER trg_track_sched_rotation_state_changes
    AFTER UPDATE ON rotation_state
    FOR EACH ROW
    WHEN(OLD.rotation_id IS NOT NULL OR NEW.rotation_id IS NOT NULL)
    EXECUTE FUNCTION fn_track_schedule_updates();

-- +migrate Down
