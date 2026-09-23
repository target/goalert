-- +migrate Up
-- A custom setting that was set with set_config(..., TRUE) does not revert to
-- NULL when the transaction ends, it reverts to an empty string for the rest of
-- the session. The previous guard checked IS NOT NULL, so once a connection had
-- queued a job for a given ID, every later transaction on that same connection
-- would silently skip queuing a job for that ID.
CREATE OR REPLACE FUNCTION fn_util_river_job (queue_name text, kind text, id text, args jsonb)
    RETURNS void
    AS $$
DECLARE
    key_name text := 'local.job__' || replace(queue_name, '-', '_') || '__' || replace(id, '-', '_');
BEGIN
    IF current_setting(key_name, TRUE) = 'true' THEN
        RETURN;
    END IF;
    -- Mark this ID as processed in this transaction
    PERFORM
        set_config(key_name, 'true', TRUE);
    INSERT INTO river_job (queue, kind, args, max_attempts, priority)
        VALUES (queue_name, kind, args, 25, 2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', jsonb_build_object('queue', queue_name)::text);
END;
$$
LANGUAGE plpgsql;

-- +migrate Down
CREATE OR REPLACE FUNCTION fn_util_river_job (queue_name text, kind text, id text, args jsonb)
    RETURNS void
    AS $$
DECLARE
    key_name text := 'local.job__' || replace(queue_name, '-', '_') || '__' || replace(id, '-', '_');
BEGIN
    IF current_setting(key_name, TRUE) IS NOT NULL THEN
        RETURN;
    END IF;
    -- Mark this ID as processed in this transaction
    PERFORM
        set_config(key_name, 'true', TRUE);
    INSERT INTO river_job (queue, kind, args, max_attempts, priority)
        VALUES (queue_name, kind, args, 25, 2);
    PERFORM
        pg_notify(current_schema() || '.river_insert', jsonb_build_object('queue', queue_name)::text);
END;
$$
LANGUAGE plpgsql;
