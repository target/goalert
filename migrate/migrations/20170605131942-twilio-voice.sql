
-- +migrate Up
CREATE TABLE twilio_voice_callbacks( 
    phone_number     TEXT NOT NULL, -- the phone number to which call was made
    callback_id      UUID NOT NULL REFERENCES sent_notifications (id) ON DELETE CASCADE, -- the unique ID of the notification 
    code             INT NOT NULL, -- the alert number 
    description      TEXT NOT NULL DEFAULT '', -- the alert description
    twilio_sid       TEXT NOT NULL, -- the unique Twilio sid for the call returned from Twilio (returned when call not delivered by Twilio)
    PRIMARY KEY (phone_number,twilio_sid)
);
-- +migrate Down
DROP TABLE IF EXISTS twilio_voice_callbacks;

