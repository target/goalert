-- +migrate Up
CREATE TABLE message_status_history(
    id bigserial PRIMARY KEY,
    message_id uuid NOT NULL REFERENCES outgoing_messages(id) ON DELETE CASCADE,
    timestamp timestamptz NOT NULL DEFAULT now(),
    status enum_outgoing_messages_status NOT NULL,
    status_details text NOT NULL
);

CREATE FUNCTION fn_insert_message_status_history()
    RETURNS TRIGGER
    AS $$
BEGIN
    INSERT INTO message_status_history(message_id, status, timestamp, status_details)
        VALUES(NEW.id, NEW.last_status, NEW.last_status_at, NEW.status_details);
    RETURN new;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_update_message_status_history
    AFTER UPDATE OF last_status ON outgoing_messages
    FOR EACH ROW
    EXECUTE FUNCTION fn_insert_message_status_history();

CREATE TRIGGER trg_insert_message_status_history
    AFTER INSERT ON outgoing_messages
    FOR EACH ROW
    EXECUTE FUNCTION fn_insert_message_status_history();

-- +migrate Down
DROP TRIGGER trg_insert_message_status_history ON outgoing_messages;

DROP TRIGGER trg_update_message_status_history ON outgoing_messages;

DROP FUNCTION fn_insert_message_status_history();

DROP TABLE message_status_history;

