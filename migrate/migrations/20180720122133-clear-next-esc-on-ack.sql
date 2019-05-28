
-- +migrate Up

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_clear_next_esc_on_alert_ack() RETURNS TRIGGER AS
$$
BEGIN

    UPDATE escalation_policy_state
    SET next_escalation = null
    WHERE alert_id = NEW.id;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd

CREATE TRIGGER trg_20_clear_next_esc_on_alert_ack
AFTER UPDATE ON public.alerts
FOR EACH ROW
WHEN (new.status != old.status and old.status = 'active'::enum_alert_status)
EXECUTE PROCEDURE fn_clear_next_esc_on_alert_ack();

-- +migrate Down

DROP TRIGGER trg_20_clear_next_esc_on_alert_ack ON alerts;
DROP FUNCTION fn_clear_next_esc_on_alert_ack();
