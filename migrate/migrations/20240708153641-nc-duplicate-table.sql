-- +migrate Up
-- Table may have not been created during a bad migration for those tracking the nightly build.
CREATE TABLE IF NOT EXISTS notification_channel_duplicates(
    old_id uuid PRIMARY KEY,
    new_id uuid NOT NULL,
    old_created_at timestamp with time zone NOT NULL
);

ALTER TABLE notification_channel_duplicates
    ADD COLUMN id bigserial UNIQUE;

-- +migrate Down
ALTER TABLE notification_channel_duplicates
    DROP COLUMN id;

