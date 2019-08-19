-- +migrate Up
INSERT INTO config_limits (id, max)
VALUES
	('notification_rules_per_user', 15),
	('contact_methods_per_user', 10),
	('ep_steps_per_policy', 10),
	('ep_actions_per_step', 20),
	('participants_per_rotation', 50),
	('rules_per_schedule', 30),
	('integration_keys_per_service', 30),
	('unacked_alerts_per_service', 200),
	('targets_per_schedule', 10),
	('heartbeat_monitors_per_service', 30),
	('user_overrides_per_schedule', 35)
ON CONFLICT DO NOTHING;

-- +migrate Down