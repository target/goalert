-- +migrate Up
UPDATE engine_processing_versions
SET "version" = 2
WHERE type_id = 'verify';

-- add new columns
ALTER TABLE user_verification_codes
    ADD COLUMN contact_method_id UUID REFERENCES user_contact_methods (id) ON DELETE CASCADE UNIQUE,
    ADD COLUMN sent boolean DEFAULT FALSE;

-- add new data (1 row for each contact method type)
UPDATE user_verification_codes code
SET
    sent = send_to IS NULL,
    contact_method_id = COALESCE(send_to, (
        SELECT contact_method_id FROM outgoing_messages
        WHERE user_verification_code_id = code.id
        LIMIT 1
    ));

DELETE FROM user_verification_codes
    WHERE contact_method_id IS NULL;

ALTER TABLE user_verification_codes
    DROP COLUMN user_id,
    DROP COLUMN contact_method_value,
    DROP COLUMN send_to,
    ALTER COLUMN sent SET NOT NULL,
    ALTER COLUMN contact_method_id SET NOT NULL;

-- +migrate Down
UPDATE engine_processing_versions
SET "version" = 1
WHERE type_id = 'verify';

ALTER TABLE user_verification_codes
    ADD COLUMN user_id UUID REFERENCES users(id),
    ADD COLUMN contact_method_value text,
    ADD COLUMN send_to UUID REFERENCES user_contact_methods(id),
    ADD CONSTRAINT user_verification_codes_user_id_contact_method_value_key UNIQUE (user_id, contact_method_value);

UPDATE user_verification_codes
SET
    user_id = cm.user_id,
    contact_method_value = cm.value,
    send_to = CASE WHEN sent THEN NULL ELSE contact_method_id END
FROM user_contact_methods cm
WHERE cm.id = contact_method_id;

ALTER TABLE user_verification_codes
    DROP COLUMN contact_method_id,
    DROP COLUMN sent,
    ALTER COLUMN user_id SET NOT NULL,
    ALTER COLUMN contact_method_value SET NOT NULL;
