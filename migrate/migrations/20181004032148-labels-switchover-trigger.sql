
-- +migrate Up notransaction

-- +migrate StatementBegin
BEGIN;
DROP TRIGGER IF EXISTS zz_99_labels_change_log ON labels;
CREATE TRIGGER zz_99_labels_change_log
AFTER INSERT OR UPDATE OR DELETE ON labels
FOR EACH ROW EXECUTE PROCEDURE process_change();
COMMIT;
-- +migrate StatementEnd

-- +migrate Down notransaction

DROP TRIGGER IF EXISTS zz_99_labels_change_log ON labels;
