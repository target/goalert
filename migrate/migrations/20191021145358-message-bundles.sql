-- +migrate Up

ALTER TABLE twilio_sms_callbacks
    ALTER alert_id DROP NOT NULL,
    ADD service_id UUID REFERENCES services (id) ON DELETE CASCADE;

CREATE INDEX idx_twilio_sms_service_id ON twilio_sms_callbacks (service_id);

ALTER TABLE outgoing_messages
    ADD status_count INT,
    ADD CONSTRAINT om_status_bundle_count CHECK (message_type <> 'alert_status_update_bundle' OR status_count NOTNULL);

-- +migrate Down

ALTER TABLE twilio_sms_callbacks DROP service_id;
DELETE FROM twilio_sms_callbacks WHERE alert_id ISNULL;
ALTER TABLE twilio_sms_callbacks ALTER alert_id SET NOT NULL;
