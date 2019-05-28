
-- +migrate Up

DROP TABLE twilio_egress_sms_status;
DROP TYPE enum_twilio_sms_status;

-- +migrate Down


CREATE TYPE enum_twilio_sms_status AS ENUM (
	'unknown', -- in case twilio insists it doesn't exist when we ask
	'accepted',
	'queued',
	'sending',
	'sent',
	'receiving',
	'received',
	'delivered',
	'undelivered',
	'failed'
);


CREATE TABLE twilio_egress_sms_status (
	twilio_sid TEXT PRIMARY KEY,
	last_status enum_twilio_sms_status NOT NULL,
	sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	last_update TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	dest_number TEXT NOT NULL
);
