-- +migrate Up

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

    IF OLD.rotation_id NOT IN (
       SELECT id FROM rotations
    ) THEN
        DELETE FROM rotation_state
        WHERE rotation_id = OLD.rotation_id;
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


-- +migrate Down

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

    IF (select 1 from rotations where id = OLD.rotation_id) != 1 THEN
        DELETE FROM rotation_state
        WHERE rotation_id = OLD.rotation_id;
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

