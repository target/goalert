
-- +migrate Up

LOCK rotation_participants;

DROP TRIGGER trg_20_decr_rot_part_position_on_delete ON rotation_participants;
DROP TRIGGER trg_30_advance_or_end_rot_on_part_del ON rotation_participants;



CREATE TRIGGER trg_20_decr_rot_part_position_on_delete AFTER DELETE ON rotation_participants FOR EACH ROW EXECUTE PROCEDURE fn_decr_rot_part_position_on_delete();
CREATE TRIGGER trg_30_advance_or_end_rot_on_part_del AFTER DELETE ON rotation_participants FOR EACH ROW EXECUTE PROCEDURE fn_advance_or_end_rot_on_part_del();


-- +migrate Down

LOCK rotation_participants;

DROP TRIGGER trg_20_decr_rot_part_position_on_delete ON rotation_participants;
DROP TRIGGER trg_30_advance_or_end_rot_on_part_del ON rotation_participants;

CREATE TRIGGER trg_20_decr_rot_part_position_on_delete BEFORE DELETE ON rotation_participants FOR EACH ROW EXECUTE PROCEDURE fn_decr_rot_part_position_on_delete();
CREATE TRIGGER trg_30_advance_or_end_rot_on_part_del BEFORE DELETE ON rotation_participants FOR EACH ROW EXECUTE PROCEDURE fn_advance_or_end_rot_on_part_del();
