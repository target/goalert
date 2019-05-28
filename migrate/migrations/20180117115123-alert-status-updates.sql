
-- +migrate Up

ALTER TABLE users
    ADD COLUMN alert_status_log_contact_method_id UUID
        REFERENCES user_contact_methods (id) ON DELETE SET NULL;

ALTER TABLE outgoing_messages
    ADD COLUMN alert_log_id BIGINT REFERENCES alert_logs (id) ON DELETE CASCADE,
    ADD CONSTRAINT om_status_update_log_id CHECK(
        message_type != 'alert_status_update' OR
        alert_log_id NOTNULL
    );

CREATE TABLE user_last_alert_log (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    log_id BIGINT NOT NULL REFERENCES alert_logs (id) ON DELETE CASCADE,
    next_log_id BIGINT NOT NULL REFERENCES alert_logs (id) ON DELETE CASCADE,

    PRIMARY KEY (user_id, alert_id)
);

-- +migrate StatementBegin
CREATE FUNCTION fn_insert_user_last_alert_log() RETURNS trigger AS $$
BEGIN

    INSERT INTO user_last_alert_log (user_id, alert_id, log_id, next_log_id)
    VALUES (NEW.sub_user_id, NEW.alert_id, NEW.id, NEW.id)
    ON CONFLICT DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION fn_update_user_last_alert_log() RETURNS trigger AS $$
BEGIN

    UPDATE user_last_alert_log last
    SET next_log_id = NEW.id
    WHERE
        last.alert_id = NEW.alert_id AND
        NEW.id > last.next_log_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER trg_insert_alert_logs_user_last_alert
AFTER INSERT
ON alert_logs
FOR EACH ROW
WHEN (NEW.event = 'notification_sent')
EXECUTE PROCEDURE fn_insert_user_last_alert_log();

CREATE TRIGGER trg_insert_alert_logs_user_last_alert_update
AFTER INSERT
ON alert_logs
FOR EACH ROW
WHEN (NEW.event IN ('acknowledged', 'closed'))
EXECUTE PROCEDURE fn_update_user_last_alert_log();

CREATE INDEX idx_alert_logs_alert_event ON alert_logs (alert_id, event);

-- +migrate Down

ALTER TABLE users
    DROP COLUMN alert_status_log_contact_method_id;

DELETE FROM outgoing_messages WHERE message_type = 'alert_status_update';

ALTER TABLE outgoing_messages
    DROP COLUMN alert_log_id;

DROP TABLE user_last_alert_log;

DROP INDEX idx_alert_logs_alert_event;

DROP TRIGGER trg_insert_alert_logs_user_last_alert_update ON alert_logs;
DROP TRIGGER trg_insert_alert_logs_user_last_alert ON alert_logs;

DROP FUNCTION fn_update_user_last_alert_log();
DROP FUNCTION fn_insert_user_last_alert_log();
