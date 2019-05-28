
-- +migrate Up



-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_status_update_same_user() RETURNS trigger AS $$
DECLARE
    _cm_user_id UUID;
BEGIN
    IF NEW.alert_status_log_contact_method_id ISNULL THEN
        RETURN NEW;
    END IF;

    SELECT INTO _cm_user_id user_id
    FROM user_contact_methods
    WHERE id = NEW.alert_status_log_contact_method_id;

    IF NEW.id != _cm_user_id THEN
        RAISE 'wrong user_id' USING ERRCODE='check_violation', CONSTRAINT='alert_status_user_id_match';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION fn_notification_rule_same_user() RETURNS trigger AS $$
DECLARE
    _cm_user_id UUID;
BEGIN
    SELECT INTO _cm_user_id user_id
    FROM user_contact_methods
    WHERE id = NEW.contact_method_id;

    IF NEW.user_id != _cm_user_id THEN
        RAISE 'wrong user_id' USING ERRCODE='check_violation', CONSTRAINT='notification_rule_user_id_match';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE TRIGGER trg_enforce_status_update_same_user 
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_status_update_same_user();


LOCK user_notification_rules;

CREATE TRIGGER trg_notification_rule_same_user 
    BEFORE INSERT OR UPDATE ON user_notification_rules
    FOR EACH ROW EXECUTE PROCEDURE fn_notification_rule_same_user();

DELETE FROM user_notification_rules r
USING user_contact_methods cm
WHERE cm.id = r.contact_method_id AND cm.user_id != r.user_id;

-- +migrate Down

DROP TRIGGER IF EXISTS trg_notification_rule_same_user ON user_notification_rules;
DROP TRIGGER IF EXISTS trg_enforce_status_update_same_user ON users;

DROP FUNCTION IF EXISTS fn_notification_rule_same_user();
DROP FUNCTION IF EXISTS fn_enforce_status_update_same_user();
