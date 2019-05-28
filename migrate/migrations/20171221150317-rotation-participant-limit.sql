
-- +migrate Up

CREATE INDEX idx_participant_rotation ON rotation_participants (rotation_id);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_rotation_participant_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'participants_per_rotation';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM rotation_participants
    WHERE rotation_id = NEW.rotation_id;

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='participants_per_rotation_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE CONSTRAINT TRIGGER trg_enforce_rotation_participant_limit 
    AFTER INSERT ON rotation_participants
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_rotation_participant_limit();

-- +migrate Down

DROP TRIGGER trg_enforce_rotation_participant_limit ON rotation_participants;
DROP FUNCTION fn_enforce_rotation_participant_limit();
DROP INDEX idx_participant_rotation;
