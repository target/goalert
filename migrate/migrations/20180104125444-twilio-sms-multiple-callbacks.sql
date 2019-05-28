
-- +migrate Up
ALTER TABLE twilio_sms_callbacks
    DROP CONSTRAINT twilio_sms_callbacks_phone_number_twilio_sid_key,
    DROP CONSTRAINT twilio_sms_callbacks_pkey;

-- +migrate Down
ALTER TABLE twilio_sms_callbacks
    ADD CONSTRAINT twilio_sms_callbacks_phone_number_twilio_sid_key UNIQUE (phone_number, twilio_sid),
    ADD CONSTRAINT twilio_sms_callbacks_pkey PRIMARY KEY (phone_number, code);

    