-- +migrate Up
LOCK notification_channels;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION nc_type_val_to_dest(typeName enum_notif_channel_type, value text)
    RETURNS jsonb
    AS $$
BEGIN
    IF typeName = 'SLACK' THEN
        RETURN jsonb_build_object('Type', 'builtin-slack-channel', 'Args', jsonb_build_object('slack_channel_id', value));
    ELSIF typeName = 'WEBHOOK' THEN
        RETURN jsonb_build_object('Type', 'builtin-webhook', 'Args', jsonb_build_object('webhook_url', value));
    ELSIF typeName = 'SLACK_USER_GROUP' THEN
        RETURN jsonb_build_object('Type', 'builtin-slack-usergroup', 'Args', jsonb_build_object('slack_usergroup_id', split_part(value, ':', 1), 'slack_channel_id', split_part(value, ':', 2)));
    ELSE
        -- throw an error
        RAISE EXCEPTION 'Unknown notification channel type: %', typeName;
    END IF;
END;
$$
LANGUAGE plpgsql;

-- +migrate StatementEnd
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_nc_compat_set_type_val_on_insert()
    RETURNS TRIGGER
    AS $$
BEGIN
    IF NEW.dest ->> 'Type' = 'builtin-slack-channel' THEN
        NEW.type = 'SLACK';
        NEW.value = NEW.dest -> 'Args' ->> 'slack_channel_id';
    ELSIF NEW.dest ->> 'Type' = 'builtin-slack-usergroup' THEN
        NEW.type = 'SLACK_USER_GROUP';
        NEW.value = NEW.dest -> 'Args' ->> 'slack_usergroup_id' || ':' || NEW.dest -> 'Args' ->> 'slack_channel_id';
    ELSIF NEW.dest ->> 'Type' = 'builtin-webhook' THEN
        NEW.type = 'WEBHOOK';
        NEW.value = NEW.dest -> 'Args' ->> 'webhook_url';
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
    notification_channels
SET
    dest = jsonb_set(dest, '{Args}',(
            SELECT
                jsonb_object_agg(replace(key, '-', '_'), value)
            FROM jsonb_each_text(dest -> 'Args')));

-- +migrate Down
LOCK notification_channels;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION nc_type_val_to_dest(typeName enum_notif_channel_type, value text)
    RETURNS jsonb
    AS $$
BEGIN
    IF typeName = 'SLACK' THEN
        RETURN jsonb_build_object('Type', 'builtin-slack-channel', 'Args', jsonb_build_object('slack-channel-id', value));
    ELSIF typeName = 'WEBHOOK' THEN
        RETURN jsonb_build_object('Type', 'builtin-webhook', 'Args', jsonb_build_object('webhook-url', value));
    ELSIF typeName = 'SLACK_USER_GROUP' THEN
        RETURN jsonb_build_object('Type', 'builtin-slack-usergroup', 'Args', jsonb_build_object('slack-usergroup-id', split_part(value, ':', 1), 'slack-channel-id', split_part(value, ':', 2)));
    ELSE
        -- throw an error
        RAISE EXCEPTION 'Unknown notification channel type: %', typeName;
    END IF;
END;
$$
LANGUAGE plpgsql;

-- +migrate StatementEnd
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_nc_compat_set_type_val_on_insert()
    RETURNS TRIGGER
    AS $$
BEGIN
    IF NEW.dest ->> 'Type' = 'builtin-slack-channel' THEN
        NEW.type = 'SLACK';
        NEW.value = NEW.dest -> 'Args' ->> 'slack-channel-id';
    ELSIF NEW.dest ->> 'Type' = 'builtin-slack-usergroup' THEN
        NEW.type = 'SLACK_USER_GROUP';
        NEW.value = NEW.dest -> 'Args' ->> 'slack-usergroup-id' || ':' || NEW.dest -> 'Args' ->> 'slack-channel-id';
    ELSIF NEW.dest ->> 'Type' = 'builtin-webhook' THEN
        NEW.type = 'WEBHOOK';
        NEW.value = NEW.dest -> 'Args' ->> 'webhook-url';
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
    notification_channels
SET
    dest = jsonb_set(dest, '{Args}',(
            SELECT
                jsonb_object_agg(replace(key, '_', '-'), value)
            FROM jsonb_each_text(dest -> 'Args')));

