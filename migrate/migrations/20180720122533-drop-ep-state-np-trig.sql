
-- +migrate Up
DROP TRIGGER trg_clear_np_cycles_on_state_delete ON escalation_policy_state;
DROP FUNCTION fn_clear_np_cycles_on_state_delete();

-- +migrate Down

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_clear_np_cycles_on_state_delete() RETURNS TRIGGER AS
$$
BEGIN
    DELETE FROM notification_policy_cycles
    WHERE alert_id = OLD.alert_id;

    RETURN OLD;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_clear_np_cycles_on_state_delete
AFTER DELETE ON public.escalation_policy_state
FOR EACH ROW
EXECUTE PROCEDURE fn_clear_np_cycles_on_state_delete();
