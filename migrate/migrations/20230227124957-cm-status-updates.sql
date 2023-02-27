-- +migrate Up
-- add enable_status_updates column to user_contact_methods table.
ALTER TABLE user_contact_methods
ADD column enable_status_updates BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE user_contact_methods
SET enable_status_updates = TRUE
WHERE id IN (
        SELECT alert_status_log_contact_method_id
        FROM users
        WHERE alert_status_log_contact_method_id IS NOT NULL
    );

-- +migrate Down
-- remove enable_status_updates column from user_contact_methods table.
ALTER TABLE user_contact_methods DROP COLUMN enable_status_updates;
