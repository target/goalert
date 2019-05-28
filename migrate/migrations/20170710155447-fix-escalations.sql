
-- +migrate Up

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
                AND (e.repeat = -1 OR (escalation_level+1) / lvl.levels <= e.repeat);
    END;
    $$;
-- +migrate StatementEnd

-- +migrate Down
