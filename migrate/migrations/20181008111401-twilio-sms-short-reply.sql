
-- +migrate Up
UPDATE engine_processing_versions
SET version = 4
WHERE type_id = 'message';

ALTER TABLE twilio_sms_callbacks
    ADD COLUMN sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    ADD COLUMN alert_id BIGINT REFERENCES alerts (id) ON DELETE CASCADE,
    DROP COLUMN twilio_sid;

CREATE INDEX idx_twilio_sms_alert_id ON twilio_sms_callbacks (alert_id);

-- cleanup old codes
DELETE FROM twilio_sms_callbacks
WHERE code NOT IN
    (
        SELECT id
        FROM alerts
        WHERE
            status != 'closed' OR
            created_at > now() - '1 day'::interval
    );

-- cleanup duplicate codes
DELETE FROM twilio_sms_callbacks
WHERE id NOT IN (
    SELECT max(id) max_id
    FROM twilio_sms_callbacks
    GROUP BY phone_number, code
);


UPDATE twilio_sms_callbacks
SET alert_id = code;

ALTER TABLE twilio_sms_callbacks
    ALTER COLUMN alert_id SET NOT NULL;

CREATE UNIQUE INDEX idx_twilio_sms_codes ON twilio_sms_callbacks (phone_number, code);

-- +migrate Down
UPDATE engine_processing_versions
SET version = 3
WHERE type_id = 'message';

ALTER TABLE twilio_sms_callbacks
    DROP COLUMN sent_at,
    DROP COLUMN alert_id,
    ADD COLUMN twilio_sid TEXT NOT NULL;

DROP INDEX idx_twilio_sms_codes;
