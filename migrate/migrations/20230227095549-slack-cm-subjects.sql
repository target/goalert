-- +migrate Up
-- add cm_id column to auth_subjects table to track the user's contact method
ALTER TABLE auth_subjects
ADD COLUMN IF NOT EXISTS cm_id UUID REFERENCES user_contact_methods (id) ON
DELETE CASCADE;

-- +migrate Down
ALTER TABLE auth_subjects DROP COLUMN IF EXISTS cm_id;
