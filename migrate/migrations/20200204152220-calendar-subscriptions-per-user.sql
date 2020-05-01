-- +migrate Up

INSERT INTO config_limits (id, max)
VALUES
	('calendar_subscriptions_per_user', 15)
ON CONFLICT DO NOTHING;


-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_calendar_subscriptions_per_user_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'calendar_subscriptions_per_user';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM user_calendar_subscriptions
    WHERE user_id = NEW.user_id;

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='calendar_subscriptions_per_user_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE CONSTRAINT TRIGGER trg_enforce_calendar_subscriptions_per_user_limit 
    AFTER INSERT ON user_calendar_subscriptions
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_calendar_subscriptions_per_user_limit();
-- +migrate StatementEnd

-- +migrate Down

DROP TRIGGER trg_enforce_calendar_subscriptions_per_user_limit ON user_calendar_subscriptions;
DROP FUNCTION fn_enforce_calendar_subscriptions_per_user_limit();
DELETE FROM config_limits WHERE id = 'calendar_subscriptions_per_user';
