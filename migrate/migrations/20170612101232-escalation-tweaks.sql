
-- +migrate Up

ALTER TABLE escalation_policy_step
    ALTER COLUMN id SET DEFAULT gen_random_uuid()::TEXT;
ALTER TABLE escalation_policy_step
    ALTER COLUMN delay SET NOT NULL;
ALTER TABLE escalation_policy_step
    ALTER COLUMN delay SET DEFAULT 1;

ALTER TABLE escalation_policy_step ADD UNIQUE(step_number, escalation_policy_id);


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION set_escalation_policy_step() RETURNS TRIGGER AS
    $$
        BEGIN
            SELECT count(step_number)+1 INTO NEW.step_number FROM escalation_policy_step WHERE escalation_policy_id = NEW.escalation_policy_id;
            RETURN NEW;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION incr_escalation_policy_steps_on_delete() RETURNS TRIGGER AS
    $$
        BEGIN
            UPDATE escalation_policy_step
            SET step_number = step_number-1
            WHERE escalation_policy_id = OLD.escalation_policy_id
                AND step_number > OLD.step_number;

            RETURN OLD;
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER incr_escalation_policy_steps_on_delete
    AFTER DELETE ON escalation_policy_step
    FOR EACH ROW
    EXECUTE PROCEDURE incr_escalation_policy_steps_on_delete();


CREATE TRIGGER set_escalation_policy_step_on_insert
    BEFORE INSERT ON escalation_policy_step
    FOR EACH ROW
    EXECUTE PROCEDURE set_escalation_policy_step();

-- +migrate Down

DROP FUNCTION set_escalation_policy_step();
