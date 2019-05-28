
-- +migrate Up

LOCK alerts, escalation_policy_state;

ALTER TABLE escalation_policy_state
    ALTER COLUMN last_escalation DROP NOT NULL,
    ALTER COLUMN last_escalation DROP DEFAULT;

INSERT INTO escalation_policy_state (alert_id, service_id, escalation_policy_id)
SELECT a.id, a.service_id, svc.escalation_policy_id
FROM alerts a
JOIN services svc ON svc.id = a.service_id
JOIN escalation_policies ep ON ep.id = svc.escalation_policy_id AND ep.step_count > 0
WHERE a.status != 'closed'
ON CONFLICT (alert_id) DO NOTHING;


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_insert_ep_state_on_step_insert() RETURNS TRIGGER AS
$$
BEGIN

    INSERT INTO escalation_policy_state (alert_id, service_id, escalation_policy_id)
    SELECT a.id, a.service_id, NEW.escalation_policy_id
    FROM alerts a
    JOIN services svc ON
        svc.id = a.service_id AND
        svc.escalation_policy_id = NEW.escalation_policy_id
    WHERE a.status != 'closed';

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_insert_ep_state_on_alert_insert() RETURNS TRIGGER AS
$$
BEGIN

    INSERT INTO escalation_policy_state (alert_id, service_id, escalation_policy_id)
    SELECT NEW.id, NEW.service_id, svc.escalation_policy_id
    FROM services svc
    JOIN escalation_policies ep ON ep.id = svc.escalation_policy_id AND ep.step_count > 0
    WHERE svc.id = NEW.service_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_10_insert_ep_state_on_alert_insert
AFTER INSERT ON public.alerts
FOR EACH ROW
WHEN (new.status != 'closed'::enum_alert_status)
EXECUTE PROCEDURE fn_insert_ep_state_on_alert_insert();

CREATE TRIGGER trg_10_insert_ep_state_on_step_insert
AFTER INSERT ON public.escalation_policy_steps
FOR EACH ROW
WHEN (NEW.step_number = 0)
EXECUTE PROCEDURE fn_insert_ep_state_on_step_insert();


-- +migrate Down
LOCK alerts, escalation_policy_state;

DELETE FROM escalation_policy_state
WHERE last_escalation isnull;

ALTER TABLE escalation_policy_state
    ALTER COLUMN last_escalation SET NOT NULL,
    ALTER COLUMN last_escalation SET DEFAULT now();

DROP TRIGGER trg_10_insert_ep_state_on_alert_insert ON public.alerts;
DROP FUNCTION fn_insert_ep_state_on_alert_insert();

DROP TRIGGER trg_10_insert_ep_state_on_step_insert ON public.escalation_policy_steps;
DROP FUNCTION fn_insert_ep_state_on_step_insert();
