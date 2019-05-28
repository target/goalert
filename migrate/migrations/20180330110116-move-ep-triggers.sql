
-- +migrate Up

LOCK escalation_policy_steps;
DROP TRIGGER trg_decr_ep_step_number_on_delete ON escalation_policy_steps;
CREATE TRIGGER trg_decr_ep_step_number_on_delete
AFTER DELETE ON escalation_policy_steps
FOR EACH ROW
EXECUTE PROCEDURE fn_decr_ep_step_number_on_delete();

-- +migrate Down
LOCK escalation_policy_steps;
DROP TRIGGER trg_decr_ep_step_number_on_delete ON escalation_policy_steps;
CREATE TRIGGER trg_decr_ep_step_number_on_delete
BEFORE DELETE ON escalation_policy_steps
FOR EACH ROW
EXECUTE PROCEDURE fn_decr_ep_step_number_on_delete();
