-- +migrate Up
CREATE OR REPLACE FUNCTION fn_enforce_signals_per_service_limit()
    RETURNS TRIGGER
    AS $$
DECLARE
    max_count int := - 1;
    val_count int := 0;
BEGIN
    SELECT
        INTO max_count max
    FROM
        config_limits
    WHERE
        id = 'pending_signals_per_service';
    IF max_count = - 1 THEN
        RETURN NEW;
    END IF;
    SELECT
        INTO val_count COUNT(*)
    FROM
        pending_signals
    WHERE
        service_id = NEW.service_id
        AND message_id IS NULL;
    IF val_count > max_count THEN
        RAISE 'limit exceeded'
        USING ERRCODE = 'check_violation', CONSTRAINT = 'pending_signals_per_service_limit', HINT = 'max=' || max_count;
    END IF;
        RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_enforce_signals_per_dest_per_service_limit()
    RETURNS TRIGGER
    AS $$
DECLARE
    max_count int := - 1;
    val_count int := 0;
BEGIN
    SELECT
        INTO max_count max
    FROM
        config_limits
    WHERE
        id = 'pending_signals_per_dest_per_service';
    IF max_count = - 1 THEN
        RETURN NEW;
    END IF;
    SELECT
        INTO val_count COUNT(*)
    FROM
        pending_signals
    WHERE
        service_id = NEW.service_id
        AND dest_id = NEW.dest_id
        AND message_id IS NULL;
    IF val_count > max_count THEN
        RAISE 'limit exceeded'
        USING ERRCODE = 'check_violation', CONSTRAINT = 'pending_signals_per_dest_per_service_limit', HINT = 'max=' || max_count;
    END IF;
        RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- +migrate Down
CREATE OR REPLACE FUNCTION fn_enforce_signals_per_service_limit()
    RETURNS TRIGGER
    AS $$
DECLARE
    max_count int := - 1;
    val_count int := 0;
BEGIN
    SELECT
        INTO max_count max
    FROM
        config_limits
    WHERE
        id = 'pending_signals_per_service';
    IF max_count = - 1 THEN
        RETURN NEW;
    END IF;
    SELECT
        INTO val_count COUNT(*)
    FROM
        pending_signals
    WHERE
        service_id = NEW.service_id;
    IF val_count > max_count THEN
        RAISE 'limit exceeded'
        USING ERRCODE = 'check_violation', CONSTRAINT = 'pending_signals_per_service_limit', HINT = 'max=' || max_count;
    END IF;
        RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION fn_enforce_signals_per_dest_per_service_limit()
    RETURNS TRIGGER
    AS $$
DECLARE
    max_count int := - 1;
    val_count int := 0;
BEGIN
    SELECT
        INTO max_count max
    FROM
        config_limits
    WHERE
        id = 'pending_signals_per_dest_per_service';
    IF max_count = - 1 THEN
        RETURN NEW;
    END IF;
    SELECT
        INTO val_count COUNT(*)
    FROM
        pending_signals
    WHERE
        service_id = NEW.service_id
        AND dest_id = NEW.dest_id;
    IF val_count > max_count THEN
        RAISE 'limit exceeded'
        USING ERRCODE = 'check_violation', CONSTRAINT = 'pending_signals_per_dest_per_service_limit', HINT = 'max=' || max_count;
    END IF;
        RETURN NEW;
END;
$$
LANGUAGE plpgsql;

