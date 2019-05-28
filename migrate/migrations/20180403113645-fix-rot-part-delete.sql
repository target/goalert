-- +migrate Up

LOCK rotation_participants;



-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_advance_or_end_rot_on_part_del() RETURNS TRIGGER AS
$$
DECLARE
    new_part UUID;
    active_part UUID;
BEGIN

    SELECT rotation_participant_id
    INTO active_part
    FROM rotation_state
    WHERE rotation_id = OLD.rotation_id;

    IF active_part != OLD.id THEN
        RETURN OLD;
    END IF;

    SELECT id
    INTO new_part
    FROM rotation_participants
    WHERE
        rotation_id = OLD.rotation_id AND
        id != OLD.id AND
        position IN (0, OLD.position+1)
    ORDER BY position DESC
    LIMIT 1;

    IF new_part ISNULL THEN
        DELETE FROM rotation_state
        WHERE rotation_id = OLD.rotation_id;
    ELSE
        UPDATE rotation_state
        SET rotation_participant_id = new_part
        WHERE rotation_id = OLD.rotation_id;
    END IF;

    RETURN OLD;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


DROP TRIGGER trg_30_advance_or_end_rot_on_part_del ON rotation_participants;

CREATE TRIGGER trg_30_advance_or_end_rot_on_part_del BEFORE DELETE ON rotation_participants FOR EACH ROW EXECUTE PROCEDURE fn_advance_or_end_rot_on_part_del();


-- +migrate Down

LOCK rotation_participants;



-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_advance_or_end_rot_on_part_del() RETURNS TRIGGER AS
$$
DECLARE
    new_part UUID;
    active_part UUID;
BEGIN

    SELECT rotation_participant_id
    INTO active_part
    FROM rotation_state
    WHERE rotation_id = OLD.rotation_id;

    IF active_part != OLD.id THEN
        RETURN OLD;
    END IF;

    SELECT id
    INTO new_part
    FROM rotation_participants
    WHERE
        rotation_id = OLD.rotation_id AND
        id != OLD.id AND
        position IN (0, OLD.position)
    ORDER BY position DESC
    LIMIT 1;

    IF new_part ISNULL THEN
        DELETE FROM rotation_state
        WHERE rotation_id = OLD.rotation_id;
    ELSE
        UPDATE rotation_state
        SET rotation_participant_id = new_part
        WHERE rotation_id = OLD.rotation_id;
    END IF;

    RETURN OLD;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


DROP TRIGGER trg_30_advance_or_end_rot_on_part_del ON rotation_participants;

CREATE TRIGGER trg_30_advance_or_end_rot_on_part_del AFTER DELETE ON rotation_participants FOR EACH ROW EXECUTE PROCEDURE fn_advance_or_end_rot_on_part_del();
