
-- +migrate Up

-- disable re-opening alerts
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION fn_prevent_reopen()
  RETURNS trigger AS
$BODY$
    BEGIN
        IF OLD.status = 'closed' THEN
            RAISE EXCEPTION 'cannot change status of closed alert';
        END IF;
        RETURN NEW;
    END;
$BODY$
  LANGUAGE plpgsql VOLATILE;
-- +migrate StatementEnd

CREATE TRIGGER trg_prevent_reopen
  BEFORE UPDATE OF status
  ON alerts
  FOR EACH ROW
  EXECUTE PROCEDURE fn_prevent_reopen();



-- collect EP snapshots when alerts are generated
CREATE TABLE alert_escalation_policy_snapshots (
    alert_id BIGINT NOT NULL REFERENCES alerts (id) ON DELETE CASCADE,
    step_number INT NOT NULL,
    step_max INT NOT NULL,
    step_delay INTERVAL NOT NULL,
    repeat INT NOT NULL,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE,
    schedule_id UUID REFERENCES schedules (id) ON DELETE CASCADE
);

CREATE VIEW alert_escalation_policies AS
    WITH step_max AS (
        SELECT escalation_policy_id, count(step_number) as step_max
        FROM escalation_policy_steps
        GROUP BY escalation_policy_id
    )
    SELECT a.id as alert_id, step.step_number, m.step_max, (step.delay::TEXT||' minutes')::INTERVAL as step_delay, e.repeat, act.user_id, act.schedule_id
    FROM
        alerts a,
        escalation_policies e,
        escalation_policy_steps step,
        step_max m,
        escalation_policy_actions act,
        services svc
    WHERE a.service_id = svc.id
        AND e.id = svc.escalation_policy_id
        AND step.escalation_policy_id = m.escalation_policy_id
        AND step.escalation_policy_id = svc.escalation_policy_id
        AND act.escalation_policy_step_id = step.id;


INSERT INTO alert_escalation_policy_snapshots
    (alert_id, step_number, step_max, step_delay, repeat, user_id, schedule_id)
SELECT alert_id, step_number, step_max, step_delay, repeat, user_id, schedule_id
FROM alert_escalation_policies;

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



-- Use snapshots when calculating notifications

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION escalate_alerts() RETURNS VOID AS
    $$
        BEGIN
            UPDATE alerts
            SET escalation_level = escalation_level + 1, last_escalation = now()
            FROM alert_escalation_policy_snapshots e
            WHERE (last_escalation + e.step_delay) < now()
                AND status = 'triggered'
                AND id = e.alert_id
                AND e.step_number = (escalation_level % e.step_max)
                AND (e.repeat = -1 OR (escalation_level+1) / e.step_max <= e.repeat);
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


CREATE OR REPLACE VIEW on_call_alert_users AS
    WITH alert_users AS (
        SELECT s.user_id,
            s.schedule_id,
            s.alert_id,
            a.status,
            a.escalation_level
        FROM
            alerts a,
            alert_escalation_policy_snapshots s
        WHERE s.alert_id = a.id
            AND s.step_number = (a.escalation_level % s.step_max)
            AND a.status <> 'closed'
    )
    SELECT DISTINCT au.alert_id,
        au.status,
        CASE
            WHEN au.user_id IS NULL THEN oc.user_id
            ELSE au.user_id
        END AS user_id,
        au.escalation_level
    FROM alert_users au
    LEFT JOIN on_call oc ON au.schedule_id = oc.schedule_id;

DROP VIEW alert_escalation_levels;

-- +migrate Down


CREATE VIEW alert_escalation_levels AS
    SELECT
        alerts.id AS alert_id,
        count(step.id) AS levels,
        alerts.escalation_level::bigint % count(step.id) AS relative_level
    FROM
        alerts,
        escalation_policy_steps step,
        services s
    WHERE step.escalation_policy_id = s.escalation_policy_id
        AND s.id = alerts.service_id
    GROUP BY alerts.id;

-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION escalate_alerts() RETURNS VOID AS
    $$
        BEGIN
            UPDATE alerts a
                SET escalation_level = escalation_level + 1, last_escalation = now()
                FROM services s, escalation_policy_steps step, alert_escalation_levels lvl, escalation_policies e
                WHERE (last_escalation + (step.delay::TEXT||' minutes')::interval) < now()
                    AND a.status = 'triggered'::enum_alert_status
                    AND s.id = a.service_id
                    AND step.escalation_policy_id = s.escalation_policy_id
                    AND lvl.alert_id = a.id
                    AND step.step_number = lvl.relative_level
                    AND e.id = s.escalation_policy_id
                    AND (e.repeat = -1 OR (escalation_level+1) / lvl.levels <= e.repeat);
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd


CREATE OR REPLACE VIEW on_call_alert_users AS
    WITH alert_users AS (
        SELECT act.user_id,
            act.schedule_id,
            a.id AS alert_id,
            a.status,
            a.escalation_level
        FROM alerts a,
            services s,
            alert_escalation_levels lvl,
            escalation_policy_steps step,
            escalation_policy_actions act
        WHERE s.id = a.service_id
            AND lvl.alert_id = a.id
            AND step.escalation_policy_id = s.escalation_policy_id
            AND step.step_number = lvl.relative_level
            AND a.status <> 'closed'::enum_alert_status
            AND act.escalation_policy_step_id = step.id
        GROUP BY act.user_id, act.schedule_id, a.id
    )
    SELECT DISTINCT au.alert_id,
        au.status,
        CASE
            WHEN au.user_id IS NULL THEN oc.user_id
            ELSE au.user_id
        END AS user_id,
        au.escalation_level
    FROM alert_users au
    LEFT JOIN on_call oc ON au.schedule_id = oc.schedule_id;

DROP TRIGGER trg_snapshot_escalation_policy ON alerts;
DROP TRIGGER trg_prevent_reopen ON alerts;
DROP FUNCTION fn_snapshot_escalation_policy();
DROP FUNCTION fn_prevent_reopen();
DROP TABLE alert_escalation_policy_snapshots;
DROP VIEW alert_escalation_policies;
