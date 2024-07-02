-- +migrate Up
CREATE TABLE pending_signals(
    id serial PRIMARY KEY,
    message_id uuid UNIQUE REFERENCES outgoing_messages(id) ON DELETE CASCADE,
    dest_id uuid NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    service_id uuid NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    params jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

ALTER TABLE pending_signals SET (autovacuum_vacuum_scale_factor = 0.01, -- More aggressive for frequent turnover
autovacuum_vacuum_threshold = 5, -- Lower threshold due to high turnover
autovacuum_analyze_scale_factor = 0.01, autovacuum_analyze_threshold = 5);

-- create index for counting signals per service and per dest per service
CREATE INDEX idx_pending_signals_service_dest_id ON pending_signals(service_id, dest_id);

INSERT INTO config_limits(id, max)
    VALUES ('pending_signals_per_service', 50),
('pending_signals_per_dest_per_service', 5)
ON CONFLICT
    DO NOTHING;

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_signals_per_service_limit()
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

-- +migrate StatementEnd
-- +migrate StatementBegin
CREATE CONSTRAINT TRIGGER trg_enforce_signals_per_service_limit
    AFTER INSERT ON pending_signals
    FOR EACH ROW
    EXECUTE PROCEDURE fn_enforce_signals_per_service_limit();

-- +migrate StatementEnd
-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_signals_per_dest_per_service_limit()
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

-- +migrate StatementEnd
-- +migrate StatementBegin
CREATE CONSTRAINT TRIGGER trg_enforce_signals_per_dest_per_service_limit
    AFTER INSERT ON pending_signals
    FOR EACH ROW
    EXECUTE PROCEDURE fn_enforce_signals_per_dest_per_service_limit();

-- +migrate StatementEnd
-- +migrate Down
DROP TABLE pending_signals;

DROP FUNCTION fn_enforce_signals_per_dest_per_service_limit();

DROP FUNCTION fn_enforce_signals_per_service_limit();

DELETE FROM config_limits
WHERE id IN ('pending_signals_per_service', 'pending_signals_per_dest_per_service');

