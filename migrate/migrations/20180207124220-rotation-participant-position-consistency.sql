
-- +migrate Up


-- read/write lock the table so once unlocked it will be both fixed and triggers in place to ensure future consistency.
LOCK rotation_participants;

-- Fix any existing discrepancies
UPDATE rotation_participants part
SET position = computed.position
FROM (
	SELECT
        id,
        row_number() OVER (PARTITION BY rotation_id ORDER BY position) - 1 AS position
	FROM rotation_participants
) computed
WHERE
    part.id = computed.id AND
    part.position != computed.position;


-- +migrate StatementBegin
CREATE FUNCTION fn_inc_rot_part_position_on_insert() RETURNS trigger AS $$
BEGIN
    LOCK rotation_participants IN EXCLUSIVE MODE;

    SELECT count(*)
    INTO NEW.position
    FROM rotation_participants
    WHERE rotation_id = NEW.rotation_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION fn_decr_rot_part_position_on_delete() RETURNS trigger AS $$
BEGIN
    LOCK rotation_participants IN EXCLUSIVE MODE;

    UPDATE rotation_participants
    SET position = position - 1
    WHERE
        rotation_id = OLD.rotation_id AND
        position > OLD.position;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_rot_part_position_no_gaps() RETURNS trigger AS $$
DECLARE
    max_pos INT := -1;
    part_count INT := 0;
BEGIN
    IF NEW.rotation_id != OLD.rotation_id THEN
        RAISE 'must not change rotation_id of existing participant';
    END IF;

    SELECT max(position), count(*)
    INTO max_pos, part_count
    FROM rotation_participants
    WHERE rotation_id = NEW.rotation_id;

    IF max_pos >= part_count THEN
        RAISE 'must not have gap in participant positions';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- ensure updates don't cause gaps
CREATE CONSTRAINT TRIGGER trg_enforce_rot_part_position_no_gaps
    AFTER UPDATE
    ON rotation_participants
    INITIALLY DEFERRED
    FOR EACH ROW
    EXECUTE PROCEDURE fn_enforce_rot_part_position_no_gaps();


CREATE TRIGGER trg_inc_rot_part_position_on_insert
    BEFORE INSERT
    ON rotation_participants
    FOR EACH ROW
    EXECUTE PROCEDURE fn_inc_rot_part_position_on_insert();

CREATE TRIGGER trg_decr_rot_part_position_on_delete
    BEFORE DELETE
    ON rotation_participants
    FOR EACH ROW
    EXECUTE PROCEDURE fn_decr_rot_part_position_on_delete();

-- +migrate Down

DROP TRIGGER trg_enforce_rot_part_position_no_gaps ON rotation_participants;
DROP FUNCTION fn_enforce_rot_part_position_no_gaps();

DROP TRIGGER trg_decr_rot_part_position_on_delete ON rotation_participants;
DROP TRIGGER trg_inc_rot_part_position_on_insert ON rotation_participants;
DROP FUNCTION fn_inc_rot_part_position_on_insert();
DROP FUNCTION fn_decr_rot_part_position_on_delete();
