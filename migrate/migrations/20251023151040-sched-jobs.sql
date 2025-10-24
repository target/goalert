-- +migrate Up
CREATE OR REPLACE FUNCTION fn_job_schedule_update(id uuid)
    RETURNS VOID
    AS $$
BEGIN
    PERFORM
        fn_util_river_job('schedule-manager'::text, 'schedule-manager-update'::text, id::text, jsonb_build_object('ScheduleID', id::text));
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_track_schedule_updates()
    RETURNS TRIGGER
    AS $$
DECLARE
    schedule_id_param uuid;
BEGIN
    IF TG_TABLE_NAME = 'schedules' THEN
        schedule_id_param := coalesce(NEW.id, OLD.id);
    ELSIF TG_TABLE_NAME = 'user_overrides' THEN
        schedule_id_param := coalesce(NEW.tgt_schedule_id, OLD.tgt_schedule_id);
    ELSE
        schedule_id_param := coalesce(NEW.schedule_id, OLD.schedule_id);
    END IF;
    PERFORM
        fn_job_schedule_update(schedule_id_param);
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_track_schedule_updates
    AFTER INSERT OR UPDATE ON schedules
    FOR EACH ROW
    EXECUTE FUNCTION fn_track_schedule_updates();

CREATE TRIGGER trg_track_schedule_rule_updates
    AFTER INSERT OR UPDATE OR DELETE ON schedule_rules
    FOR EACH ROW
    EXECUTE FUNCTION fn_track_schedule_updates();

CREATE TRIGGER trg_track_schedule_data_updates
    AFTER INSERT OR UPDATE OR DELETE ON schedule_data
    FOR EACH ROW
    EXECUTE FUNCTION fn_track_schedule_updates();

CREATE TRIGGER trg_track_schedule_user_override_updates
    AFTER INSERT OR UPDATE OR DELETE ON user_overrides
    FOR EACH ROW
    EXECUTE FUNCTION fn_track_schedule_updates();

CREATE OR REPLACE FUNCTION fn_track_rot_schedule_updates()
    RETURNS TRIGGER
    AS $$
BEGIN
    PERFORM
        fn_job_schedule_update(rule.schedule_id)
    FROM
        schedule_rules rule
    WHERE
        rule.tgt_rotation_id = coalesce(NEW.rotation_id, OLD.rotation_id);
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_track_rot_schedule_updates
    AFTER INSERT OR UPDATE OR DELETE ON rotation_state
    FOR EACH ROW
    EXECUTE FUNCTION fn_track_rot_schedule_updates();

SELECT
    fn_job_schedule_update(id)
FROM
    schedules;

-- +migrate Down
DROP TRIGGER trg_track_schedule_user_override_updates ON user_overrides;

DROP TRIGGER trg_track_schedule_data_updates ON schedule_data;

DROP TRIGGER trg_track_schedule_rule_updates ON schedule_rules;

DROP TRIGGER trg_track_schedule_updates ON schedules;

DROP TRIGGER trg_track_rot_schedule_updates ON rotation_state;

DROP FUNCTION fn_track_schedule_updates();

DROP FUNCTION fn_track_rot_schedule_updates();

DROP FUNCTION fn_job_schedule_update(id uuid);

DELETE FROM river_job
WHERE kind = 'schedule-manager-update';

