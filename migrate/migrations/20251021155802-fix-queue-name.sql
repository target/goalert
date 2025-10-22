-- +migrate Up
CREATE OR REPLACE FUNCTION fn_util_river_job(queue_name text, kind text, id
    text, args jsonb)
    RETURNS void
    AS $$
DECLARE
    key_name text := 'local.job__' || replace(queue_name, '-',
	'_') || '__' || replace(id, '-',
	'_');
BEGIN
    IF current_setting(key_name, TRUE) IS NOT NULL THEN
        RETURN;
    END IF;
    -- Mark this ID as processed in this transaction
    PERFORM
        set_config(key_name, 'true', TRUE);
    INSERT INTO river_job(
        queue,
        kind,
        args,
        max_attempts,
        priority)
    VALUES (
        queue_name,
        kind,
        args,
        25,
        2);
    PERFORM
	pg_notify(current_schema() || '.river_insert',
	    jsonb_build_object('queue', queue_name)::text);
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_track_alert_status()
    RETURNS TRIGGER
    AS $$
BEGIN
    PERFORM
	fn_util_river_job('status-manager'::text, 'status-manager-look-for-work'::text,
	    NEW.id::text, jsonb_build_object('AlertID', NEW.id));
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_job_rotation(id uuid)
    RETURNS VOID
    AS $$
BEGIN
    PERFORM
	fn_util_river_job('rotation-manager'::text, 'rotation-manager-update'::text,
	    id::text, jsonb_build_object('RotationID', id));
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_job_signal()
    RETURNS TRIGGER
    AS $$
BEGIN
    PERFORM
	fn_util_river_job('engine-signal-mgr'::text, 'signal-manager-schedule-outgoing-messages'::text,
	    NEW.service_id::text, jsonb_build_object('ServiceID',
	    NEW.service_id::text));
    RETURN NULL;
END;
$$
LANGUAGE plpgsql;

-- +migrate Down
CREATE OR REPLACE FUNCTION fn_track_alert_status()
    RETURNS TRIGGER
    AS $$
BEGIN
    INSERT INTO river_job(
        args,
        kind,
        max_attempts,
        priority)
    VALUES(
        json_build_object(
            'AlertID', NEW.id),
        'status-manager-look-for-work',
        25,
        2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', '{"queue":"status-manager"}');
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_job_rotation(id uuid)
    RETURNS VOID
    AS $$
DECLARE
    key_name text := 'rotation_job_' || replace(id::text, '-', '_');
BEGIN
    -- Only run once per transaction per ID
    IF current_setting('local.' || key_name, TRUE) IS NOT NULL THEN
        RETURN;
    END IF;
    -- Mark this ID as processed in this transaction
    PERFORM
        set_config('local.' || key_name, 'true', TRUE);
    INSERT INTO river_job(
        args,
        kind,
        max_attempts,
        priority)
    VALUES (
        json_build_object(
            'RotationID', id),
        'rotation-manager-update',
        25,
        2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', '{"queue":"rotation-manager"}');
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_job_signal()
    RETURNS TRIGGER
    AS $$
DECLARE
    key_name text := 'signal_job_' || replace(NEW.service_id::text,
	'-', '_');
BEGIN
    -- Only run once per transaction per ID
    IF current_setting('local.' || key_name, TRUE) IS NOT NULL THEN
        RETURN NULL;
    END IF;
    -- Mark this ID as processed in this transaction
    PERFORM
        set_config('local.' || key_name, 'true', TRUE);
    INSERT INTO river_job(
        args,
        kind,
        max_attempts,
        priority)
    VALUES (
        json_build_object(
            'ServiceID', NEW.service_id),
        'signal-manager-schedule-outgoing-messages',
        25,
        2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', '{"queue":"engine-signal-mgr"}');
    RETURN NULL;
END;
$$
LANGUAGE plpgsql;

DROP FUNCTION IF EXISTS fn_util_river_job(text, text, text, jsonb);
