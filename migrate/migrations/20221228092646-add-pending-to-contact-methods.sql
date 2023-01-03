-- +migrate Up

BEGIN;

-- Add pending column with default to false for existing records and update default to true for new records later on
ALTER TABLE user_contact_methods ADD COLUMN pending BOOLEAN NOT NULL DEFAULT FALSE;

-- Update default to true for new records
ALTER TABLE user_contact_methods ALTER COLUMN pending SET DEFAULT TRUE;

-- Add a function to set pending to false when disabled is set to false for compatibility with previous versions
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION set_pending_to_false() RETURNS TRIGGER AS $$
BEGIN
  IF NEW.disabled = FALSE THEN
    NEW.pending = FALSE;
  END IF;
  RETURN NEW;
END;
$$ language plpgsql;
-- +migrate StatementEnd

-- Bind the trigger function to user_contact_methods
CREATE TRIGGER set_pending_to_false
BEFORE UPDATE OF disabled ON user_contact_methods
FOR EACH ROW EXECUTE PROCEDURE set_pending_to_false();

COMMIT;

-- +migrate Down

BEGIN;

ALTER TABLE user_contact_methods DROP COLUMN pending;

DROP TRIGGER set_pending_to_false ON user_contact_methods;

DROP FUNCTION set_pending_to_false();

COMMIT;
