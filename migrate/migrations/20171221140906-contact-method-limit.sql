
-- +migrate Up

CREATE INDEX idx_contact_method_users ON user_contact_methods (user_id);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_contact_method_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'contact_methods_per_user';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM user_contact_methods
    WHERE user_id = NEW.user_id;

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='contact_methods_per_user_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE CONSTRAINT TRIGGER trg_enforce_contact_method_limit 
    AFTER INSERT ON user_contact_methods
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_contact_method_limit();

-- +migrate Down

DROP TRIGGER trg_enforce_contact_method_limit ON user_contact_methods;
DROP FUNCTION fn_enforce_contact_method_limit();
DROP INDEX idx_contact_method_users;
