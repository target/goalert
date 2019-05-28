
-- +migrate Up

CREATE INDEX idx_rule_schedule ON schedule_rules (schedule_id);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_schedule_rule_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'rules_per_schedule';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM schedule_rules
    WHERE schedule_id = NEW.schedule_id;

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='rules_per_schedule_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE CONSTRAINT TRIGGER trg_enforce_schedule_rule_limit 
    AFTER INSERT ON schedule_rules
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_schedule_rule_limit();

-- +migrate Down

DROP TRIGGER trg_enforce_schedule_rule_limit ON schedule_rules;
DROP FUNCTION fn_enforce_schedule_rule_limit();
DROP INDEX idx_rule_schedule;
