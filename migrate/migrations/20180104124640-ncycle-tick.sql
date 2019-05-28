
-- +migrate Up

LOCK
	notification_policy_cycles,
	notification_logs;

ALTER TABLE notification_policy_cycles
	ADD COLUMN last_tick TIMESTAMP WITH TIME ZONE;


-- add sent messages for active cycles
INSERT INTO outgoing_messages (
	id,
	message_type,
	contact_method_id,
	created_at,
	last_status,
	last_status_at,
	status_details,
	sent_at,
	alert_id,
	cycle_id,
	user_id,
	service_id,
	escalation_policy_id
)
SELECT
	log.id,
	cast('alert_notification' as enum_outgoing_messages_type),
	log.contact_method_id,
	log.process_timestamp,
	cast(case when log.completed then 'sent' else 'pending' end as enum_outgoing_messages_status),
	log.process_timestamp,
	'migrated',
	case when log.completed then process_timestamp else null end,
	log.alert_id,
	cycle.id,
	cm.user_id,
	a.service_id,
	svc.escalation_policy_id
FROM notification_logs log
JOIN notification_policy_cycles cycle ON cycle.alert_id = log.alert_id AND cycle.checked AND cycle.started_at <= log.process_timestamp
JOIN user_contact_methods cm ON cm.id = log.contact_method_id AND cm.user_id = cycle.user_id
JOIN alerts a ON a.id = log.alert_id
JOIN services svc ON svc.id = a.service_id
ORDER BY process_timestamp DESC
ON CONFLICT DO NOTHING
;


with last_sent as (
	select distinct
		alert_id,
		cm.user_id,
		max(process_timestamp)
	from notification_logs log
	join user_contact_methods cm on cm.id = log.contact_method_id
	where log.completed
	group by alert_id, cm.user_id
)
update notification_policy_cycles cycle
set last_tick = last_sent.max
from last_sent
where
	last_sent.alert_id = cycle.alert_id and
	last_sent.user_id = cycle.user_id;

-- +migrate Down

DELETE FROM outgoing_messages WHERE status_details = 'migrated';
ALTER TABLE notification_policy_cycles
	DROP COLUMN last_tick;
