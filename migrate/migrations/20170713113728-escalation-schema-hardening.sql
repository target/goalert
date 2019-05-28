
-- +migrate Up

UPDATE escalation_policy SET description = '' WHERE description IS NULL;
UPDATE escalation_policy SET repeat = 0 WHERE repeat IS NULL;

ALTER TABLE escalation_policy ALTER COLUMN description SET DEFAULT '';
ALTER TABLE escalation_policy ALTER COLUMN description SET NOT NULL;

ALTER TABLE escalation_policy ALTER COLUMN repeat SET DEFAULT 0;
ALTER TABLE escalation_policy ALTER COLUMN repeat SET NOT NULL;


UPDATE service SET name = '' WHERE name IS NULL;
UPDATE service SET description = '' WHERE description IS NULL;

ALTER TABLE service ALTER COLUMN name SET NOT NULL;

ALTER TABLE service ALTER COLUMN description SET DEFAULT '';
ALTER TABLE service ALTER COLUMN description SET NOT NULL;

-- +migrate Down

ALTER TABLE escalation_policy
    ALTER repeat DROP NOT NULL,
    ALTER repeat DROP DEFAULT,
    ALTER description DROP NOT NULL,
    ALTER description DROP DEFAULT;

ALTER TABLE service
    ALTER name DROP NOT NULL,
    ALTER name DROP DEFAULT,
    ALTER description DROP NOT NULL,
    ALTER description DROP DEFAULT;
