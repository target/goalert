
-- +migrate Up
DROP TABLE twilio_voice_callbacks;

-- +migrate Down
CREATE TABLE twilio_voice_callbacks (
    phone_number text,
    callback_id uuid NOT NULL,
    code integer NOT NULL,
    description text NOT NULL DEFAULT ''::text,
    twilio_sid text,
    CONSTRAINT twilio_voice_callbacks_pkey PRIMARY KEY (phone_number, twilio_sid)
);
