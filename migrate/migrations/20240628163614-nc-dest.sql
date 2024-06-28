-- +migrate Up
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
ALTER TABLE notification_channels
    ADD COLUMN dest jsonb UNIQUE;

UPDATE
    notification_channels
SET
    dest = nc_type_val_to_dest(type, value);

ALTER TABLE notification_channels
    ALTER COLUMN dest SET NOT NULL;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_nc_set_dest_on_insert()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.dest = nc_type_val_to_dest(NEW.type, NEW.value);
    RETURN new;
END;
$$
LANGUAGE plpgsql;

-- +migrate StatementEnd
CREATE TRIGGER trg_10_nc_set_dest_on_insert
    BEFORE INSERT ON notification_channels
    FOR EACH ROW
    WHEN(NEW.dest IS NULL)
    EXECUTE FUNCTION fn_nc_set_dest_on_insert();

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
CREATE TRIGGER trg_10_nc_compat_set_type_val_on_insert
    BEFORE INSERT ON notification_channels
    FOR EACH ROW
    WHEN(NEW.dest IS NOT NULL)
    EXECUTE FUNCTION fn_nc_compat_set_type_val_on_insert();

-- +migrate Down
ALTER TABLE notification_channels
    DROP COLUMN dest;

DELETE FROM notification_channels
WHERE type = 'DEST';

DROP TRIGGER trg_10_nc_compat_set_type_val_on_insert ON notification_channels;

DROP FUNCTION fn_nc_compat_set_type_val_on_insert();

DROP TRIGGER trg_10_nc_set_dest_on_insert ON notification_channels;

DROP FUNCTION fn_nc_set_dest_on_insert();

DROP FUNCTION nc_type_val_to_dest(typeName enum_notif_channel_type, value text);

