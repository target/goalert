
-- +migrate Up

UPDATE escalation_policy_step SET step_number = step_number - 1;


-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION escalate_alerts() RETURNS void
    LANGUAGE plpgsql
    AS $$
    BEGIN
        UPDATE alerts a
            SET escalation_level = escalation_level + 1, last_escalation = now()
            FROM service s, escalation_policy_step step, alert_escalation_levels lvl, escalation_policy e
            WHERE (last_escalation + (step.delay::TEXT||' minutes')::interval) < now()
                AND a.status = 'triggered'::enum_alert_status
                AND s.id = a.service_id
                AND step.escalation_policy_id = s.escalation_policy_id
                AND lvl.alert_id = a.id
                AND step.step_number = (a.escalation_level % lvl.levels)
                AND e.id = s.escalation_policy_id
                AND (e.repeat = -1 OR escalation_level / lvl.levels < e.repeat);
    END;
    $$;
-- +migrate StatementEnd

-- +migrate StatementBegin

CREATE OR REPLACE FUNCTION set_escalation_policy_step() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
        BEGIN
            SELECT count(step_number) INTO NEW.step_number FROM escalation_policy_step WHERE escalation_policy_id = NEW.escalation_policy_id;
            RETURN NEW;
        END;
    $$;
-- +migrate StatementEnd

CREATE OR REPLACE VIEW alert_escalation_levels AS
    SELECT alerts.id AS alert_id, 
            count(step.id) AS levels,
            (alerts.escalation_level % count(step.id)) as relative_level 
        FROM alerts,escalation_policy_step step,service
        WHERE step.escalation_policy_id = service.escalation_policy_id
            AND service.id = alerts.service_id
        GROUP BY alerts.id;


SELECT update_notifications();

-- +migrate Down
