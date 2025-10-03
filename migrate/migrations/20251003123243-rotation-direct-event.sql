-- +migrate Up
CREATE OR REPLACE FUNCTION fn_job_rotation(id uuid)
    RETURNS VOID
    AS $$
DECLARE
    key_name text := 'rotation_job_' || id::text;
BEGIN
    -- Only run once per transaction per ID
    IF current_setting('local.' || key_name, TRUE) IS NOT NULL THEN
        RETURN;
    END IF;
    -- Mark this ID as processed in this transaction
    PERFORM
        set_config('local.' || key_name, 'true', TRUE);
    INSERT INTO river_job(args, kind, max_attempts, priority)
        VALUES (json_build_object('RotationID', id), 'rotation-manager-update', 25, 2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', '{"queue":"rotation-manager"}');
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_track_rotation_updates()
    RETURNS TRIGGER
    AS $$
BEGIN
    IF TG_TABLE_NAME = 'rotations' THEN
        PERFORM
            fn_job_rotation(NEW.id);
    ELSIF TG_OP = 'DELETE' THEN
        PERFORM
            fn_job_rotation(OLD.rotation_id);
        RETURN OLD;
    ELSE
        PERFORM
            fn_job_rotation(NEW.rotation_id);
    END IF;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- +migrate Down
CREATE OR REPLACE FUNCTION fn_track_rotation_updates()
    RETURNS TRIGGER
    AS $$
BEGIN
    IF TG_TABLE_NAME = 'rotations' THEN
        INSERT INTO entity_updates(entity_type, entity_id)
            VALUES('rotation', NEW.id);
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO entity_updates(entity_type, entity_id)
            VALUES('rotation', OLD.rotation_id);
        RETURN OLD;
    ELSE
        INSERT INTO entity_updates(entity_type, entity_id)
            VALUES('rotation', NEW.rotation_id);
    END IF;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS fn_job_rotation(uuid);

