
-- +migrate Up

CREATE INDEX idx_unacked_alert_service ON alerts ("status", service_id);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_alert_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'unacked_alerts_per_service';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM alerts
    WHERE service_id = NEW.service_id AND "status" = 'triggered';

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='unacked_alerts_per_service_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE CONSTRAINT TRIGGER trg_enforce_alert_limit 
    AFTER INSERT ON alerts
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_alert_limit();

-- +migrate Down

DROP TRIGGER trg_enforce_alert_limit ON alerts;
DROP FUNCTION fn_enforce_alert_limit();
DROP INDEX idx_unacked_alert_service;
