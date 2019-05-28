
-- +migrate Up

ALTER TABLE escalation_policy_steps
  DROP CONSTRAINT escalation_policy_steps_escalation_policy_id_step_number_key,
  ADD UNIQUE(escalation_policy_id, step_number) DEFERRABLE INITIALLY DEFERRED; -- Needs to be deferrable in order to reorder steps else query will fail.

-- +migrate StatementBegin

CREATE OR REPLACE FUNCTION move_escalation_policy_step(_id UUID, _new_pos INT) RETURNS VOID AS
  $$
  DECLARE
    _old_pos INT;
    _epid UUID;
  BEGIN
    SELECT step_number, escalation_policy_id into _old_pos, _epid FROM escalation_policy_steps WHERE id = _id;
    IF _old_pos > _new_pos THEN
      UPDATE escalation_policy_steps
      SET step_number = step_number + 1
      WHERE escalation_policy_id = _epid
        AND step_number < _old_pos
        AND step_number >= _new_pos;
    ELSE
      UPDATE escalation_policy_steps
      SET step_number = step_number - 1
      WHERE escalation_policy_id = _epid
        AND step_number > _old_pos
        AND step_number <= _new_pos;
    END IF;
    UPDATE escalation_policy_steps
    SET step_number = _new_pos
    WHERE id = _id;
  END;
  $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd
-- +migrate Down

ALTER TABLE escalation_policy_steps
  DROP CONSTRAINT escalation_policy_steps_escalation_policy_id_step_number_key,
  ADD CONSTRAINT escalation_policy_steps_escalation_policy_id_step_number_key UNIQUE (escalation_policy_id, step_number);

DROP FUNCTION move_escalation_policy_step(_id UUID, _new_pos INT);
