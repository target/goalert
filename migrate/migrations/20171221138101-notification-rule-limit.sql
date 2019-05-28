
-- +migrate Up

CREATE INDEX idx_notification_rule_users ON user_notification_rules (user_id);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_notification_rule_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'notification_rules_per_user';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM user_notification_rules
    WHERE user_id = NEW.user_id;

    IF max_count != -1 AND val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='notification_rules_per_user_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE CONSTRAINT TRIGGER trg_enforce_notification_rule_limit 
    AFTER INSERT ON user_notification_rules
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_notification_rule_limit();

-- +migrate Down

DROP TRIGGER trg_enforce_notification_rule_limit ON user_notification_rules;
DROP FUNCTION fn_enforce_notification_rule_limit();
DROP INDEX idx_notification_rule_users;
