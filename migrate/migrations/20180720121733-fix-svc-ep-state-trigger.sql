
-- +migrate Up
LOCK services, alerts, escalation_policy_state;

DROP TRIGGER trg_reset_ep_state_on_ep_change ON escalation_policy_state;
DROP FUNCTION fn_reset_ep_state_on_ep_change();

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_clear_ep_state_on_svc_ep_change() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE escalation_policy_state
    SET
        escalation_policy_id = NEW.escalation_policy_id,
        escalation_policy_step_id = NULL,
        loop_count = 0,
        last_escalation = NULL,
        next_escalation = NULL,
        force_escalation = false,
        escalation_policy_step_number = 0
    WHERE service_id = NEW.id
    ;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

DROP TRIGGER trg_10_clear_ep_state_on_svc_ep_change ON services;
CREATE TRIGGER trg_10_clear_ep_state_on_svc_ep_change
AFTER UPDATE ON public.services
FOR EACH ROW
WHEN ((old.escalation_policy_id <> new.escalation_policy_id))
EXECUTE PROCEDURE fn_clear_ep_state_on_svc_ep_change();

-- +migrate Down

LOCK services, alerts, escalation_policy_state;

DROP TRIGGER trg_10_clear_ep_state_on_svc_ep_change ON services;
CREATE TRIGGER trg_10_clear_ep_state_on_svc_ep_change
BEFORE UPDATE ON public.services
FOR EACH ROW
WHEN ((old.escalation_policy_id <> new.escalation_policy_id))
EXECUTE PROCEDURE fn_clear_ep_state_on_svc_ep_change();

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_clear_ep_state_on_svc_ep_change() RETURNS TRIGGER AS
$$
BEGIN
    DELETE FROM escalation_policy_state
        WHERE service_id = NEW.id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_reset_ep_state_on_ep_change() RETURNS TRIGGER AS
$$
BEGIN
    
    NEW.escalation_policy_step_number = 0;
    NEW.loop_count = 0;
    NEW.force_escalation = FALSE;
    NEW.last_escalation = now();

    SELECT id INTO NEW.escalation_policy_step_id
    FROM escalation_policy_steps
    WHERE
        step_number = 0 and
        escalation_policy_id = NEW.escalation_policy_id;

    DELETE FROM notification_policy_cycles
    WHERE service_id = NEW.service_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_reset_ep_state_on_ep_change
AFTER UPDATE ON escalation_policy_state
FOR EACH ROW
WHEN (NEW.escalation_policy_id <> OLD.escalation_policy_id)
EXECUTE PROCEDURE fn_reset_ep_state_on_ep_change();
