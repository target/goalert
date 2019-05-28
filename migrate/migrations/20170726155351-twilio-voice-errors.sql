
-- +migrate Up

CREATE TABLE twilio_voice_errors (
    phone_number TEXT NOT NULL,
    error_message TEXT NOT NULL,
    outgoing BOOLEAN NOT NULL,
    occurred_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE INDEX ON twilio_voice_errors (phone_number, outgoing, occurred_at);

-- +migrate Down

DROP TABLE twilio_voice_errors;
