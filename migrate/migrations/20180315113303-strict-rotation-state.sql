
-- +migrate Up

ALTER TABLE rotation_state
    DROP CONSTRAINT rotation_state_rotation_participant_id_fkey,
    ADD CONSTRAINT rotation_state_rotation_participant_id_fkey
        FOREIGN KEY (rotation_participant_id)
        REFERENCES rotation_participants (id)
        ON DELETE RESTRICT,
    ALTER rotation_participant_id SET NOT NULL;


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_set_rot_state_pos_on_active_change() RETURNS TRIGGER AS
$$
BEGIN
    SELECT position INTO NEW.position
    FROM rotation_participants
    WHERE id = NEW.rotation_participant_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_set_rot_state_pos_on_part_reorder() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE rotation_state
    SET position = NEW.position
    WHERE rotation_participant_id = NEW.id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

ALTER TABLE rotations
    ADD COLUMN participant_count INT NOT NULL DEFAULT 0;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_incr_part_count_on_add() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE rotations
    SET participant_count = participant_count + 1
    WHERE id = NEW.rotation_id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_decr_part_count_on_del() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE rotations
    SET participant_count = participant_count - 1
    WHERE id = OLD.rotation_id;

    RETURN OLD;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_start_rotation_on_first_part_add() RETURNS TRIGGER AS
$$
DECLARE
    first_part UUID;
BEGIN
    SELECT id
    INTO first_part
    FROM rotation_participants
    WHERE rotation_id = NEW.rotation_id AND position = 0;

    INSERT INTO rotation_state (
        rotation_id, rotation_participant_id, shift_start
    ) VALUES (
        NEW.rotation_id, first_part, now()
    ) ON CONFLICT DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

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



LOCK rotation_participants;
WITH part_count AS (
    SELECT rotation_id, count(*)
    FROM rotation_participants
    GROUP BY rotation_id
)
UPDATE rotations
SET participant_count = part_count.count
FROM part_count
WHERE part_count.rotation_id = rotations.id;

INSERT INTO rotation_state (rotation_id, rotation_participant_id, shift_start)
SELECT rotation_id, id, now()
FROM rotation_participants
WHERE position = 0
ON CONFLICT (rotation_id) DO NOTHING;

CREATE TRIGGER trg_set_rot_state_pos_on_active_change
BEFORE UPDATE ON rotation_state
FOR EACH ROW
WHEN (NEW.rotation_participant_id != OLD.rotation_participant_id)
EXECUTE PROCEDURE fn_set_rot_state_pos_on_active_change();

CREATE TRIGGER trg_set_rot_state_pos_on_part_reorder
BEFORE UPDATE ON rotation_participants
FOR EACH ROW
WHEN (NEW.position != OLD.position)
EXECUTE PROCEDURE fn_set_rot_state_pos_on_part_reorder();

CREATE TRIGGER trg_incr_part_count_on_add
BEFORE INSERT ON rotation_participants
FOR EACH ROW
EXECUTE PROCEDURE fn_incr_part_count_on_add();


CREATE TRIGGER trg_start_rotation_on_first_part_add
AFTER INSERT ON rotation_participants
FOR EACH ROW
EXECUTE PROCEDURE fn_start_rotation_on_first_part_add();


CREATE TRIGGER trg_10_decr_part_count_on_del
BEFORE DELETE ON rotation_participants
FOR EACH ROW
EXECUTE PROCEDURE fn_decr_part_count_on_del();


DROP TRIGGER trg_decr_rot_part_position_on_delete ON rotation_participants;

CREATE TRIGGER trg_20_decr_rot_part_position_on_delete
BEFORE DELETE ON rotation_participants
FOR EACH ROW
EXECUTE PROCEDURE fn_decr_rot_part_position_on_delete();

CREATE TRIGGER trg_30_advance_or_end_rot_on_part_del
BEFORE DELETE ON rotation_participants
FOR EACH ROW
EXECUTE PROCEDURE fn_advance_or_end_rot_on_part_del();

-- +migrate Down

ALTER TABLE rotation_state
    ALTER rotation_participant_id DROP NOT NULL,
    DROP CONSTRAINT rotation_state_rotation_participant_id_fkey,
    ADD CONSTRAINT rotation_state_rotation_participant_id_fkey
        FOREIGN KEY (rotation_participant_id)
        REFERENCES rotation_participants (id)
        ON DELETE SET NULL;

DROP TRIGGER trg_set_rot_state_pos_on_active_change ON rotation_state;
DROP TRIGGER trg_set_rot_state_pos_on_part_reorder ON rotation_participants;
DROP TRIGGER trg_incr_part_count_on_add ON rotation_participants;
DROP TRIGGER trg_start_rotation_on_first_part_add ON rotation_participants;
DROP TRIGGER trg_10_decr_part_count_on_del ON rotation_participants;

DROP TRIGGER trg_20_decr_rot_part_position_on_delete ON rotation_participants;

CREATE TRIGGER trg_decr_rot_part_position_on_delete
BEFORE DELETE ON rotation_participants
FOR EACH ROW
EXECUTE PROCEDURE fn_decr_rot_part_position_on_delete();

DROP TRIGGER trg_30_advance_or_end_rot_on_part_del ON rotation_participants;

DROP FUNCTION fn_set_rot_state_pos_on_active_change();
DROP FUNCTION fn_set_rot_state_pos_on_part_reorder();
DROP FUNCTION fn_incr_part_count_on_add();
DROP FUNCTION fn_decr_part_count_on_del();
DROP FUNCTION fn_start_rotation_on_first_part_add();
DROP FUNCTION fn_advance_or_end_rot_on_part_del();

ALTER TABLE rotations
    DROP COLUMN participant_count;
