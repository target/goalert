
-- +migrate Up


-- read/write lock the table so once unlocked it will be both fixed and triggers in place to ensure future consistency.
LOCK escalation_policy_steps;

-- Fix any existing discrepancies
UPDATE escalation_policy_steps step
SET step_number = computed.step_number
FROM (
	SELECT
        id,
        row_number() OVER (PARTITION BY escalation_policy_id ORDER BY step_number) - 1 AS step_number
	FROM escalation_policy_steps
) computed
WHERE
    step.id = computed.id AND
    step.step_number != computed.step_number;

-- +migrate StatementBegin
CREATE FUNCTION fn_inc_ep_step_number_on_insert() RETURNS trigger AS $$
BEGIN
    LOCK escalation_policy_steps IN EXCLUSIVE MODE;

    SELECT count(*)
    INTO NEW.step_number
    FROM escalation_policy_steps
    WHERE escalation_policy_id = NEW.escalation_policy_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION fn_decr_ep_step_number_on_delete() RETURNS trigger AS $$
BEGIN
    LOCK escalation_policy_steps IN EXCLUSIVE MODE;

    UPDATE escalation_policy_steps
    SET step_number = step_number - 1
    WHERE
        escalation_policy_id = OLD.escalation_policy_id AND
        step_number > OLD.step_number;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_ep_step_number_no_gaps() RETURNS trigger AS $$
DECLARE
    max_pos INT := -1;
    step_count INT := 0;
BEGIN
    IF NEW.escalation_policy_id != OLD.escalation_policy_id THEN
        RAISE 'must not change escalation_policy_id of existing step';
    END IF;

    SELECT max(step_number), count(*)
    INTO max_pos, step_count
    FROM escalation_policy_steps
    WHERE escalation_policy_id = NEW.escalation_policy_id;

    IF max_pos >= step_count THEN
        RAISE 'must not have gap in step_numbers';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- ensure updates don't cause gaps
CREATE CONSTRAINT TRIGGER trg_ep_step_number_no_gaps
    AFTER UPDATE
    ON escalation_policy_steps
    INITIALLY DEFERRED
    FOR EACH ROW
    EXECUTE PROCEDURE fn_enforce_ep_step_number_no_gaps();

DROP TRIGGER incr_escalation_policy_steps_on_delete ON escalation_policy_steps;
DROP TRIGGER set_escalation_policy_step_on_insert ON escalation_policy_steps;
DROP FUNCTION set_escalation_policy_step();
DROP FUNCTION incr_escalation_policy_steps_on_delete();

CREATE TRIGGER trg_inc_ep_step_number_on_insert
    BEFORE INSERT
    ON escalation_policy_steps
    FOR EACH ROW
    EXECUTE PROCEDURE fn_inc_ep_step_number_on_insert();

CREATE TRIGGER trg_decr_ep_step_number_on_delete
    BEFORE DELETE
    ON escalation_policy_steps
    FOR EACH ROW
    EXECUTE PROCEDURE fn_decr_ep_step_number_on_delete();

-- +migrate Down

DROP TRIGGER trg_ep_step_number_no_gaps ON escalation_policy_steps;
DROP FUNCTION fn_enforce_ep_step_number_no_gaps();

-- +migrate StatementBegin
CREATE FUNCTION set_escalation_policy_step() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
SELECT count(step_number) INTO NEW.step_number FROM escalation_policy_steps WHERE escalation_policy_id = NEW.escalation_policy_id;
RETURN NEW;
END;
$$;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION incr_escalation_policy_steps_on_delete() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
UPDATE escalation_policy_steps
SET step_number = step_number-1
WHERE escalation_policy_id = OLD.escalation_policy_id
AND step_number > OLD.step_number;

RETURN OLD;
END;
$$;
-- +migrate StatementEnd

CREATE TRIGGER incr_escalation_policy_steps_on_delete AFTER DELETE ON escalation_policy_steps FOR EACH ROW EXECUTE PROCEDURE incr_escalation_policy_steps_on_delete();
CREATE TRIGGER set_escalation_policy_step_on_insert BEFORE INSERT ON escalation_policy_steps FOR EACH ROW EXECUTE PROCEDURE set_escalation_policy_step();
DROP TRIGGER trg_decr_ep_step_number_on_delete ON escalation_policy_steps;
DROP TRIGGER trg_inc_ep_step_number_on_insert ON escalation_policy_steps;
DROP FUNCTION fn_decr_ep_step_number_on_delete();
DROP FUNCTION fn_inc_ep_step_number_on_insert();
