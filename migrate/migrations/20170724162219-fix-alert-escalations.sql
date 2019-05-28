
-- +migrate Up

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


-- +migrate Down

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
                    AND step.step_number = ((a.escalation_level + 1) % lvl.levels)
                    AND e.id = s.escalation_policy_id
                    AND (e.repeat = -1 OR escalation_level / lvl.levels < e.repeat);
        END;
    $$ LANGUAGE 'plpgsql';
-- +migrate StatementEnd
