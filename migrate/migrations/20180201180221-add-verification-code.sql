
-- +migrate Up
CREATE TABLE user_verification_codes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    contact_method_value TEXT NOT NULL,
    code int NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    send_to UUID REFERENCES user_contact_methods(id),
    UNIQUE(user_id, contact_method_value)
);

ALTER TABLE user_contact_methods
    ADD COLUMN last_test_verify_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE outgoing_messages
    ADD COLUMN user_verification_code_id UUID REFERENCES user_verification_codes(id) ON DELETE CASCADE,
    ADD CONSTRAINT verify_needs_id CHECK(message_type != 'verification_message' OR user_verification_code_id NOTNULL);

-- +migrate Down

ALTER TABLE user_contact_methods
    DROP COLUMN last_test_verify_at;

ALTER TABLE outgoing_messages
    DROP COLUMN user_verification_code_id;

DROP TABLE user_verification_codes;
