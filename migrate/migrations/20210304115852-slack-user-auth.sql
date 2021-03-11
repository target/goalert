-- +migrate Up

ALTER TABLE user_slack_data RENAME COLUMN id TO user_id;
ALTER TABLE user_slack_data ADD COLUMN slack_id TEXT NOT NULL;
ALTER TABLE user_slack_data ADD COLUMN team_id TEXT NOT NULL;

CREATE TABLE notification_channel_last_alert_log (
    notification_channel_id UUID NOT NULL REFERENCES notification_channels (id) ON DELETE CASCADE,
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    log_id BIGINT NOT NULL REFERENCES alert_logs (id) ON DELETE CASCADE,
    next_log_id BIGINT NOT NULL REFERENCES alert_logs (id) ON DELETE CASCADE,

    PRIMARY KEY (notification_channel_id, alert_id)
);

-- +migrate Down

ALTER TABLE user_slack_data DROP COLUMN team_id;
ALTER TABLE user_slack_data DROP COLUMN slack_id;
ALTER TABLE user_slack_data RENAME COLUMN user_id TO id;

DROP TABLE notification_channel_last_alert_log;