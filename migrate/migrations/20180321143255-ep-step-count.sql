
-- +migrate Up

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_incr_ep_step_count_on_add() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE escalation_policies
    SET step_count = step_count + 1
    WHERE id = NEW.escalation_policy_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_decr_ep_step_count_on_del() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE escalation_policies
    SET step_count = step_count - 1
    WHERE id = OLD.escalation_policy_id;

    RETURN OLD;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


ALTER TABLE escalation_policies
    ADD COLUMN step_count INT NOT NULL DEFAULT 0;

LOCK escalation_policy_steps IN EXCLUSIVE MODE;

WITH counts AS (
    SELECT escalation_policy_id, count(*)
    FROM escalation_policy_steps
    GROUP BY escalation_policy_id
)
UPDATE escalation_policies
SET step_count = counts.count
FROM counts
WHERE counts.escalation_policy_id = escalation_policies.id;


CREATE TRIGGER trg_10_incr_ep_step_count_on_add
BEFORE INSERT ON escalation_policy_steps
FOR EACH ROW
EXECUTE PROCEDURE fn_incr_ep_step_count_on_add();

CREATE TRIGGER trg_10_decr_ep_step_count_on_del
BEFORE DELETE ON escalation_policy_steps
FOR EACH ROW
EXECUTE PROCEDURE fn_decr_ep_step_count_on_del();


-- +migrate Down

DROP TRIGGER trg_10_incr_ep_step_count_on_add on escalation_policy_steps;
DROP TRIGGER trg_10_decr_ep_step_count_on_del on escalation_policy_steps;

DROP FUNCTION fn_decr_ep_step_count_on_del();
DROP FUNCTION fn_incr_ep_step_count_on_add();

ALTER TABLE escalation_policies
    DROP COLUMN step_count;

