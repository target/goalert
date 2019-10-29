-- +migrate Up

ALTER TABLE twilio_sms_callbacks
    ALTER alert_id DROP NOT NULL,
    ADD service_id UUID REFERENCES services (id) ON DELETE CASCADE;

CREATE INDEX idx_twilio_sms_service_id ON twilio_sms_callbacks (service_id);

ALTER TABLE outgoing_messages
    ADD status_alert_ids BIGINT[],
    ADD CONSTRAINT om_status_alert_ids CHECK (message_type <> 'alert_status_update_bundle' OR status_alert_ids NOTNULL),
    DROP CONSTRAINT om_processed_no_fired_sent,
    ADD CONSTRAINT om_processed_no_fired_sent CHECK(
        last_status in ('pending','sending','failed','bundled') or
        (fired_at isnull and sent_at notnull)
    );

-- +migrate Down

ALTER TABLE twilio_sms_callbacks DROP service_id;
DELETE FROM twilio_sms_callbacks WHERE alert_id ISNULL;
ALTER TABLE twilio_sms_callbacks ALTER alert_id SET NOT NULL;

DELETE FROM outgoing_messages WHERE message_type IN ('alert_notification_bundle', 'alert_status_update_bundle');
ALTER TABLE outgoing_messages
    DROP status_alert_ids,
    DROP CONSTRAINT om_processed_no_fired_sent,
    ADD CONSTRAINT om_processed_no_fired_sent CHECK(
        last_status in ('pending','sending','failed') or
        (fired_at isnull and sent_at notnull)
    );
