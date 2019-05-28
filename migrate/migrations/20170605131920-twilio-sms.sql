
-- +migrate Up
CREATE TABLE twilio_sms_callbacks( 
    phone_number     TEXT NOT NULL, -- the phone number to be notified
    callback_id      UUID NOT NULL REFERENCES sent_notifications (id) ON DELETE CASCADE, -- the unique ID of the notification 
    code             INT NOT NULl, -- the alert number with which user should respond to for this alert
    twilio_sid       TEXT NOT NULL, -- the unique Twilio sid returned from Twilio (returned when alert message not delivered by Twilio)
    PRIMARY KEY(phone_number,code),
    UNIQUE (phone_number,twilio_sid)
);


-- +migrate Down
DROP TABLE IF EXISTS twilio_sms_callbacks;

