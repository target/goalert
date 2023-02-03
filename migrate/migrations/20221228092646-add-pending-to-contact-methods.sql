-- +migrate Up

-- A pending column is being added to user_contact_methods to indicate if a contact method has ever been verified.
-- This can be used for various purposes such as cleaning up unverified contact methods after a certain period of time.
-- Defaulting it to false for existing records to preserve existing behavior on existing records.
ALTER TABLE user_contact_methods ADD COLUMN pending BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE user_contact_methods ALTER COLUMN pending SET DEFAULT TRUE;

-- A contact method is only pending (i.e., subject to cleanup) until disabled is set to false (e.g., after first verification).
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_cm_set_not_pending_on_verify() RETURNS TRIGGER AS $$
BEGIN
    NEW.pending = FALSE;
  RETURN NEW;
END;
$$ language plpgsql;
-- +migrate StatementEnd

CREATE TRIGGER trg_cm_set_not_pending_on_verify
BEFORE UPDATE OF disabled ON user_contact_methods
FOR EACH ROW
WHEN (NOT NEW.disabled AND OLD.pending)
EXECUTE PROCEDURE fn_cm_set_not_pending_on_verify();

-- +migrate Down

DROP TRIGGER trg_cm_set_not_pending_on_verify ON user_contact_methods;

DROP FUNCTION fn_cm_set_not_pending_on_verify();

ALTER TABLE user_contact_methods DROP COLUMN pending;
