
-- +migrate Up
DROP TRIGGER trg_snapshot_escalation_policy ON alerts;
DROP FUNCTION fn_snapshot_escalation_policy();

-- +migrate Down

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_snapshot_escalation_policy()
  RETURNS trigger AS
$BODY$
    BEGIN
        INSERT INTO alert_escalation_policy_snapshots
            (alert_id, step_number, step_max, step_delay, repeat, user_id, schedule_id)
        SELECT alert_id, step_number, step_max, step_delay, repeat, user_id, schedule_id
        FROM alert_escalation_policies pol
        WHERE pol.alert_id = NEW.id;

        RETURN NEW;
    END;
$BODY$
  LANGUAGE plpgsql VOLATILE;
-- +migrate StatementEnd


CREATE TRIGGER trg_snapshot_escalation_policy
    AFTER INSERT
    ON alerts
    FOR EACH ROW
    EXECUTE PROCEDURE fn_snapshot_escalation_policy();