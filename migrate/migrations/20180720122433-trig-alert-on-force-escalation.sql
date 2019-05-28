
-- +migrate Up

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_lock_svc_on_force_escalation() RETURNS TRIGGER AS
$$
BEGIN

    -- lock service first
    PERFORM 1
    FROM services svc
    WHERE svc.id = NEW.service_id
    FOR UPDATE;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_trig_alert_on_force_escalation() RETURNS TRIGGER AS
$$
BEGIN

    UPDATE alerts
    SET "status" = 'triggered'
    WHERE id = NEW.alert_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


CREATE TRIGGER trg_20_lock_svc_on_force_escalation
BEFORE UPDATE ON public.escalation_policy_state
FOR EACH ROW
WHEN (new.force_escalation != old.force_escalation and new.force_escalation)
EXECUTE PROCEDURE fn_lock_svc_on_force_escalation();

CREATE TRIGGER trg_30_trig_alert_on_force_escalation
AFTER UPDATE ON public.escalation_policy_state
FOR EACH ROW
WHEN (new.force_escalation != old.force_escalation and new.force_escalation)
EXECUTE PROCEDURE fn_trig_alert_on_force_escalation();

-- +migrate Down

DROP TRIGGER trg_20_lock_svc_on_force_escalation on escalation_policy_state;
DROP TRIGGER trg_30_trig_alert_on_force_escalation on escalation_policy_state;
DROP FUNCTION fn_trig_alert_on_force_escalation();
DROP FUNCTION fn_lock_svc_on_force_escalation();