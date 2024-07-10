-- +migrate Up
LOCK user_contact_methods;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION cm_type_val_to_dest(typeName enum_user_contact_method_type, value text)
    RETURNS jsonb
    AS $$
BEGIN
    IF typeName = 'EMAIL' THEN
        RETURN jsonb_build_object('Type', 'builtin-smtp-email', 'Args', jsonb_build_object('email_address', value));
    ELSIF typeName = 'VOICE' THEN
        RETURN jsonb_build_object('Type', 'builtin-twilio-voice', 'Args', jsonb_build_object('phone_number', value));
    ELSIF typeName = 'SMS' THEN
        RETURN jsonb_build_object('Type', 'builtin-twilio-sms', 'Args', jsonb_build_object('phone_number', value));
    ELSIF typeName = 'WEBHOOK' THEN
        RETURN jsonb_build_object('Type', 'builtin-webhook', 'Args', jsonb_build_object('webhook_url', value));
    ELSIF typeName = 'SLACK_DM' THEN
        RETURN jsonb_build_object('Type', 'builtin-slack-dm', 'Args', jsonb_build_object('slack_user_id', value));
    ELSE
        -- throw an error
        RAISE EXCEPTION 'Unknown contact method type: %', typeName;
    END IF;
END;
$$
LANGUAGE plpgsql;

-- +migrate StatementEnd
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_cm_compat_set_type_val_on_insert()
    RETURNS TRIGGER
    AS $$
BEGIN
    IF NEW.dest ->> 'Type' = 'builtin-smtp-email' THEN
        NEW.type = 'EMAIL';
        NEW.value = NEW.dest -> 'Args' ->> 'email_address';
    ELSIF NEW.dest ->> 'Type' = 'builtin-twilio-voice' THEN
        NEW.type = 'VOICE';
        NEW.value = NEW.dest -> 'Args' ->> 'phone_number';
    ELSIF NEW.dest ->> 'Type' = 'builtin-twilio-sms' THEN
        NEW.type = 'SMS';
        NEW.value = NEW.dest -> 'Args' ->> 'phone_number';
    ELSIF NEW.dest ->> 'Type' = 'builtin-webhook' THEN
        NEW.type = 'WEBHOOK';
        NEW.value = NEW.dest -> 'Args' ->> 'webhook_url';
    ELSIF NEW.dest ->> 'Type' = 'builtin-slack-dm' THEN
        NEW.type = 'SLACK_DM';
        NEW.value = NEW.dest -> 'Args' ->> 'slack_user_id';
    ELSE
        NEW.type = 'DEST';
        NEW.value = gen_random_uuid()::text;
    END IF;
    RETURN new;
END;
$$
LANGUAGE plpgsql;

-- +migrate StatementEnd
UPDATE
    user_contact_methods
SET
    dest = jsonb_set(dest, '{Args}',(
            SELECT
                jsonb_object_agg(replace(key, '-', '_'), value)
            FROM jsonb_each_text(dest -> 'Args')));

-- +migrate Down
LOCK user_contact_methods;

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
UPDATE
    user_contact_methods
SET
    dest = jsonb_set(dest, '{Args}',(
            SELECT
                jsonb_object_agg(replace(key, '_', '-'), value)
            FROM jsonb_each_text(dest -> 'Args')));

