
-- +migrate Up

CREATE TABLE user_overrides (
    id UUID PRIMARY KEY,

    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    CHECK(end_time > start_time), -- needs name
    CHECK(end_time > now()),

    add_user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    remove_user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    CHECK(COALESCE(add_user_id, remove_user_id) NOTNULL),
    CHECK(add_user_id != remove_user_id),

    tgt_schedule_id UUID NOT NULL REFERENCES schedules (id) ON DELETE CASCADE
);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_user_overide_no_conflict() RETURNS trigger AS $$
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


CREATE CONSTRAINT TRIGGER trg_enforce_user_overide_no_conflict 
    AFTER INSERT OR UPDATE ON user_overrides
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_user_overide_no_conflict();

CREATE INDEX idx_user_overrides_schedule ON user_overrides (tgt_schedule_id, end_time);

-- +migrate StatementBegin
CREATE FUNCTION fn_enforce_user_override_schedule_limit() RETURNS trigger AS $$
DECLARE
    max_count INT := -1;
    val_count INT := 0;
BEGIN
    SELECT INTO max_count max
    FROM config_limits
    WHERE id = 'user_overrides_per_schedule';

    IF max_count = -1 THEN
        RETURN NEW;
    END IF;

    SELECT INTO val_count COUNT(*)
    FROM user_overrides
    WHERE
        tgt_schedule_id = NEW.tgt_schedule_id AND
        end_time > now();

    IF val_count > max_count THEN
        RAISE 'limit exceeded' USING ERRCODE='check_violation', CONSTRAINT='user_overrides_per_schedule_limit', HINT='max='||max_count;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd


CREATE CONSTRAINT TRIGGER trg_enforce_user_override_schedule_limit 
    AFTER INSERT ON user_overrides
    FOR EACH ROW EXECUTE PROCEDURE fn_enforce_user_override_schedule_limit();
    

-- +migrate Down

DROP TABLE user_overrides;

DROP FUNCTION fn_enforce_user_overide_no_conflict();
DROP FUNCTION fn_enforce_user_override_schedule_limit();
