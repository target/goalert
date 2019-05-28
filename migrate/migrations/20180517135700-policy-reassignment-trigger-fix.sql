
-- +migrate Up

LOCK escalation_policy_state;

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

ALTER TABLE services
    ADD CONSTRAINT svc_ep_uniq UNIQUE(id, escalation_policy_id);

ALTER TABLE escalation_policy_state
    ADD CONSTRAINT svc_ep_fkey
        FOREIGN KEY (service_id, escalation_policy_id)
        REFERENCES services (id, escalation_policy_id)
        ON DELETE CASCADE
        ON UPDATE CASCADE;

-- +migrate Down

ALTER TABLE escalation_policy_state
    DROP CONSTRAINT svc_ep_fkey;
ALTER TABLE services
    DROP CONSTRAINT svc_ep_uniq;


DROP TRIGGER trg_reset_ep_state_on_ep_change ON escalation_policy_state;
DROP FUNCTION fn_reset_ep_state_on_ep_change();
