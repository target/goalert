-- +migrate Up
UPDATE engine_processing_versions
SET "version" = 4
WHERE type_id = 'status_update';

-- add enable_status_updates column to user_contact_methods table.
ALTER TABLE user_contact_methods
ADD column enable_status_updates BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE user_contact_methods
SET enable_status_updates = TRUE
WHERE id IN (
        SELECT alert_status_log_contact_method_id
        FROM users
        WHERE alert_status_log_contact_method_id IS NOT NULL
    )
    OR TYPE = 'SLACK_DM'
    OR TYPE = 'WEBHOOK';

-- +migrate Down
UPDATE engine_processing_versions
SET "version" = 3
WHERE type_id = 'status_update';

-- remove enable_status_updates column from user_contact_methods table.
ALTER TABLE user_contact_methods DROP COLUMN enable_status_updates;
