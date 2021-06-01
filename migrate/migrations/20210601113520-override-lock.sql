-- +migrate Up

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_enforce_user_overide_no_conflict() RETURNS trigger AS $$
DECLARE
    conflict UUID := NULL;
BEGIN
    EXECUTE 'LOCK user_overrides IN EXCLUSIVE MODE';

    SELECT id INTO conflict
    FROM user_overrides
    WHERE
        id != NEW.id AND
        tgt_schedule_id = NEW.tgt_schedule_id AND
        (
            add_user_id in (NEW.remove_user_id, NEW.add_user_id) OR
            remove_user_id in (NEW.remove_user_id, NEW.add_user_id)
        ) AND
        (start_time, end_time) OVERLAPS (NEW.start_time, NEW.end_time)
    FOR UPDATE
    LIMIT 1;
  
    IF conflict NOTNULL THEN
        RAISE 'override conflict' USING ERRCODE='check_violation', CONSTRAINT='user_override_no_conflict_allowed', HINT='CONFLICTING_ID='||conflict::text;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_enforce_user_overide_no_conflict() RETURNS trigger AS $$
DECLARE
    conflict UUID := NULL;
BEGIN
    SELECT id INTO conflict
    FROM user_overrides
    WHERE
        id != NEW.id AND
        tgt_schedule_id = NEW.tgt_schedule_id AND
        (
            add_user_id in (NEW.remove_user_id, NEW.add_user_id) OR
            remove_user_id in (NEW.remove_user_id, NEW.add_user_id)
        ) AND
        (start_time, end_time) OVERLAPS (NEW.start_time, NEW.end_time)
    LIMIT 1;
  
    IF conflict NOTNULL THEN
        RAISE 'override conflict' USING ERRCODE='check_violation', CONSTRAINT='user_override_no_conflict_allowed', HINT='CONFLICTING_ID='||conflict::text;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd
