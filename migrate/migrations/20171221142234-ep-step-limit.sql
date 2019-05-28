
-- +migrate Up

CREATE INDEX idx_ep_step_policies ON escalation_policy_steps (escalation_policy_id);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_ep_step_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'ep_steps_per_policy';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM escalation_policy_steps
    WHERE escalation_policy_id = NEW.escalation_policy_id;

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='ep_steps_per_policy_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE CONSTRAINT TRIGGER trg_enforce_ep_step_limit 
    AFTER INSERT ON escalation_policy_steps
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_ep_step_limit();

-- +migrate Down

DROP TRIGGER trg_enforce_ep_step_limit ON escalation_policy_steps;
DROP FUNCTION fn_enforce_ep_step_limit();
DROP INDEX idx_ep_step_policies;
