
-- +migrate Up notransaction
ALTER TYPE enum_alert_log_subject_type ADD VALUE IF NOT EXISTS 'channel';

ALTER TABLE alert_logs
    ADD COLUMN sub_channel_id uuid REFERENCES notification_channels (id) ON DELETE SET NULL;
CREATE INDEX idx_alert_logs_channel_id ON alert_logs(sub_channel_id uuid_ops);

-- +migrate Down

ALTER TABLE alert_logs DROP COLUMN sub_channel_id;
