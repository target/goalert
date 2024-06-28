-- +migrate Up
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION cm_type_val_to_dest(typeName enum_user_contact_method_type, value text)
    RETURNS jsonb
    AS $$
BEGIN
    IF typeName = 'EMAIL' THEN
        RETURN jsonb_build_object('Type', 'builtin-smtp-email', 'Args', jsonb_build_object('email-address', value));
    ELSIF typeName = 'VOICE' THEN
        RETURN jsonb_build_object('Type', 'builtin-twilio-voice', 'Args', jsonb_build_object('phone-number', value));
    ELSIF typeName = 'SMS' THEN
        RETURN jsonb_build_object('Type', 'builtin-twilio-sms', 'Args', jsonb_build_object('phone-number', value));
    ELSIF typeName = 'WEBHOOK' THEN
        RETURN jsonb_build_object('Type', 'builtin-webhook', 'Args', jsonb_build_object('webhook-url', value));
    ELSIF typeName = 'SLACK_DM' THEN
        RETURN jsonb_build_object('Type', 'builtin-slack-dm', 'Args', jsonb_build_object('slack-user-id', value));
    ELSE
        -- throw an error
        RAISE EXCEPTION 'Unknown contact method type: %', typeName;
    END IF;
END;
$$
LANGUAGE plpgsql;

-- +migrate StatementEnd
ALTER TABLE user_contact_methods
    ADD COLUMN dest jsonb UNIQUE;

UPDATE
    user_contact_methods
SET
    dest = cm_type_val_to_dest(type, value);

ALTER TABLE user_contact_methods
    ALTER COLUMN dest SET NOT NULL;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_cm_set_dest_on_insert()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.dest = cm_type_val_to_dest(NEW.type, NEW.value);
    RETURN new;
END;
$$
LANGUAGE plpgsql;

-- +migrate StatementEnd
CREATE TRIGGER trg_10_cm_set_dest_on_insert
    BEFORE INSERT ON user_contact_methods
    FOR EACH ROW
    WHEN(NEW.dest IS NULL)
    EXECUTE FUNCTION fn_cm_set_dest_on_insert();

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_cm_compat_set_type_val_on_insert()
    RETURNS TRIGGER
    AS $$
BEGIN
    IF NEW.dest ->> 'Type' = 'builtin-smtp-email' THEN
        NEW.type = 'EMAIL';
        NEW.value = NEW.dest -> 'Args' ->> 'email-address';
    ELSIF NEW.dest ->> 'Type' = 'builtin-twilio-voice' THEN
        NEW.type = 'VOICE';
        NEW.value = NEW.dest -> 'Args' ->> 'phone-number';
    ELSIF NEW.dest ->> 'Type' = 'builtin-twilio-sms' THEN
        NEW.type = 'SMS';
        NEW.value = NEW.dest -> 'Args' ->> 'phone-number';
    ELSIF NEW.dest ->> 'Type' = 'builtin-webhook' THEN
        NEW.type = 'WEBHOOK';
        NEW.value = NEW.dest -> 'Args' ->> 'webhook-url';
    ELSIF NEW.dest ->> 'Type' = 'builtin-slack-dm' THEN
        NEW.type = 'SLACK_DM';
        NEW.value = NEW.dest -> 'Args' ->> 'slack-user-id';
    ELSE
        NEW.type = 'DEST';
        NEW.value = gen_random_uuid()::text;
    END IF;
    RETURN new;
END;
$$
LANGUAGE plpgsql;

-- +migrate StatementEnd
CREATE TRIGGER trg_10_compat_set_type_val_on_insert
    BEFORE INSERT ON user_contact_methods
    FOR EACH ROW
    WHEN(NEW.dest IS NOT NULL)
    EXECUTE FUNCTION fn_cm_compat_set_type_val_on_insert();

-- +migrate Down
ALTER TABLE user_contact_methods
    DROP COLUMN dest;

DELETE FROM user_contact_methods
WHERE type = 'DEST';

DROP TRIGGER trg_10_compat_set_type_val_on_insert ON user_contact_methods;

DROP FUNCTION fn_cm_compat_set_type_val_on_insert();

DROP TRIGGER trg_10_cm_set_dest_on_insert ON user_contact_methods;

DROP FUNCTION fn_cm_set_dest_on_insert();

DROP FUNCTION cm_type_val_to_dest(enum_user_contact_methods_type, text);

