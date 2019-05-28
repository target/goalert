
-- +migrate Up

CREATE INDEX idx_heartbeat_monitor_service ON heartbeat_monitors (service_id);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_heartbeat_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'heartbeat_monitors_per_service';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM heartbeat_monitors
    WHERE service_id = NEW.service_id;

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='heartbeat_monitors_per_service_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE CONSTRAINT TRIGGER trg_enforce_heartbeat_monitor_limit 
    AFTER INSERT ON heartbeat_monitors
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_heartbeat_limit();

-- +migrate Down

DROP TRIGGER trg_enforce_heartbeat_monitor_limit ON heartbeat_monitors;
DROP FUNCTION fn_enforce_heartbeat_limit();
DROP INDEX idx_heartbeat_monitor_service;
