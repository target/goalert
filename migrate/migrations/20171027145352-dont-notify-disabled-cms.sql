
-- +migrate Up

CREATE OR REPLACE VIEW needs_notification_sent AS
    SELECT DISTINCT
        cs.alert_id,
        nr.contact_method_id,
        cm.type,
        cm.value,
        a.description,
        s.name AS service_name,
        nr.id AS notification_rule_id,
        cs.escalation_level,
        cs.cycle_id
    FROM
        user_notification_cycle_state cs,
        alerts a,
        user_contact_methods cm,
        user_notification_rules nr,
        services s
    WHERE a.id = cs.alert_id
        AND a.status = 'triggered'::enum_alert_status
        AND cs.escalation_level = a.escalation_level
        AND cm.id = nr.contact_method_id
        AND nr.id = cs.notification_rule_id
        AND s.id = a.service_id
        AND cs.pending
        AND NOT cs.future
        AND cm.disabled = FALSE;

-- +migrate Down

CREATE OR REPLACE VIEW needs_notification_sent AS
    SELECT DISTINCT
        cs.alert_id,
        nr.contact_method_id,
        cm.type,
        cm.value,
        a.description,
        s.name AS service_name,
        nr.id AS notification_rule_id,
        cs.escalation_level,
        cs.cycle_id
    FROM
        user_notification_cycle_state cs,
        alerts a,
        user_contact_methods cm,
        user_notification_rules nr,
        services s
    WHERE a.id = cs.alert_id
        AND a.status = 'triggered'::enum_alert_status
        AND cs.escalation_level = a.escalation_level
        AND cm.id = nr.contact_method_id
        AND nr.id = cs.notification_rule_id
        AND s.id = a.service_id
        AND cs.pending
        AND NOT cs.future;