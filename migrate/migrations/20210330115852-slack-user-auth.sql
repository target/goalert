-- +migrate Up
ALTER TABLE user_slack_data RENAME COLUMN id TO user_id;
ALTER TABLE user_slack_data ADD COLUMN slack_id TEXT NOT NULL;
ALTER TABLE user_slack_data ADD COLUMN team_id TEXT NOT NULL;

CREATE TABLE notification_channel_last_alert_log (
    notification_channel_id UUID NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    alert_id BIGINT NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    log_id BIGINT NOT NULL REFERENCES alert_logs(id) ON DELETE CASCADE,
    next_log_id BIGINT NOT NULL REFERENCES alert_logs(id) ON DELETE CASCADE,
    PRIMARY KEY (notification_channel_id, alert_id)
);

-- +migrate StatementBegin
CREATE FUNCTION fn_insert_notification_channel_last_alert_log() RETURNS trigger AS $$
BEGIN
    INSERT INTO notification_channel_last_alert_log
        (notification_channel_id, alert_id, log_id, next_log_id)
    VALUES
        (NEW.sub_channel_id, NEW.alert_id, NEW.id, NEW.id)
    ON CONFLICT DO NOTHING;
RETURN NEW;
END;
        $$ LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER trg_insert_alert_logs_notification_channel_last_alert
AFTER INSERT ON
alert_logs
FOR EACH ROW 
WHEN (NEW.event = 'notification_sent' AND NEW.sub_type = 'channel')
EXECUTE PROCEDURE fn_insert_notification_channel_last_alert_log();

-- +migrate StatementBegin
CREATE FUNCTION fn_update_notification_channel_last_alert_log() RETURNS trigger AS $$
BEGIN
    UPDATE notification_channel_last_alert_log last
    SET next_log_id
    = NEW.id
    WHERE
        last.alert_id = NEW.alert_id AND
        NEW.id > last.next_log_id;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TRIGGER trg_insert_alert_logs_notification_channel_last_alert_update
AFTER INSERT
ON alert_logs
FOR EACH ROW
WHEN (NEW.event IN ('acknowledged', 'closed'))
EXECUTE PROCEDURE fn_update_notification_channel_last_alert_log();

-- +migrate Down
DROP TRIGGER trg_insert_alert_logs_notification_channel_last_alert_update ON alert_logs;
DROP TRIGGER trg_insert_alert_logs_notification_channel_last_alert ON alert_logs;

ALTER TABLE user_slack_data DROP COLUMN team_id;
ALTER TABLE user_slack_data DROP COLUMN slack_id;
ALTER TABLE user_slack_data RENAME COLUMN user_id TO id;

DROP FUNCTION fn_update_notification_channel_last_alert_log();
DROP FUNCTION fn_insert_notification_channel_last_alert_log();
DROP TABLE notification_channel_last_alert_log;