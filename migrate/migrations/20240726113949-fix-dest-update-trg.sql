-- +migrate Up
DROP TRIGGER trg_10_nc_set_dest_on_update ON notification_channels;

CREATE TRIGGER trg_10_nc_set_dest_on_update
    AFTER UPDATE ON notification_channels
    FOR EACH ROW
    WHEN(NEW.dest = OLD.dest AND(NEW.type != 'DEST' AND(NEW.value != OLD.value OR NEW.type != OLD.type)))
    EXECUTE FUNCTION fn_nc_set_dest_on_insert();

DROP TRIGGER trg_10_cm_set_dest_on_update ON user_contact_methods;

CREATE TRIGGER trg_10_cm_set_dest_on_update
    AFTER UPDATE ON user_contact_methods
    FOR EACH ROW
    WHEN(NEW.dest = OLD.dest AND(NEW.type != 'DEST' AND(NEW.value != OLD.value OR NEW.type != OLD.type)))
    EXECUTE FUNCTION fn_cm_set_dest_on_insert();

-- +migrate Down
DROP TRIGGER trg_10_nc_set_dest_on_update ON notification_channels;

CREATE TRIGGER trg_10_nc_set_dest_on_update
    BEFORE UPDATE ON notification_channels
    FOR EACH ROW
    WHEN(NEW.dest = OLD.dest)
    EXECUTE FUNCTION fn_nc_set_dest_on_insert();

DROP TRIGGER trg_10_cm_set_dest_on_update ON user_contact_methods;

CREATE TRIGGER trg_10_cm_set_dest_on_update
    BEFORE UPDATE ON user_contact_methods
    FOR EACH ROW
    WHEN(NEW.dest = OLD.dest)
    EXECUTE FUNCTION fn_cm_set_dest_on_insert();

