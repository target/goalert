

-- +migrate Up

CREATE TYPE enum_twilio_voice_status AS ENUM (
	'unknown', -- in case twilio insists it doesn't exist when we ask
	'initiated',
	'queued',
	'ringing',
	'in-progress',
	'completed',
	'busy',
	'failed',
	'no-answer',
	'canceled'
);


CREATE TABLE twilio_egress_voice_status (
	twilio_sid TEXT PRIMARY KEY,
	last_status enum_twilio_voice_status NOT NULL,
	sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	last_update TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	dest_number TEXT NOT NULL,
	last_sequence_number INT
);

-- +migrate Down

DROP TABLE twilio_egress_voice_status;
DROP TYPE enum_twilio_voice_status;
