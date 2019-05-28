
-- +migrate Up

ALTER TABLE rotations
    ADD COLUMN time_zone TEXT;

-- inherit timezone
UPDATE rotations
SET
    time_zone = s.time_zone,
    name = s.name||' Rotation'
FROM schedules s
WHERE s.id = schedule_id;

ALTER TABLE rotations
    ALTER COLUMN time_zone SET NOT NULL,
    ADD CONSTRAINT rotations_name_unique UNIQUE (name);

INSERT INTO schedule_rules (schedule_id, tgt_rotation_id)
SELECT schedule_id, id
FROM rotations;

ALTER TABLE rotations
    DROP COLUMN schedule_id,
    ALTER COLUMN time_zone SET NOT NULL;

-- +migrate Down

ALTER TABLE rotations
    DROP COLUMN time_zone,
    ADD COLUMN schedule_id UUID REFERENCES schedules (id) ON DELETE CASCADE,
    DROP CONSTRAINT rotations_name_unique;

UPDATE rotations rot
SET schedule_id = rule.schedule_id
FROM schedule_rules rule
WHERE rule.tgt_rotation_id = rot.id;

ALTER TABLE rotations
    ALTER COLUMN schedule_id SET NOT NULL,
    ADD CONSTRAINT rotations_schedule_id_name_key UNIQUE (schedule_id, name);

DELETE FROM schedule_rules;
