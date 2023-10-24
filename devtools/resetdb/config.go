package main

// Default values for data generation
const (
	UserCount            = 1619  // select count(id) from users
	CMMax                = 7     // select count(id) from user_contact_methods group by user_id order by count desc limit 1
	NRMax                = 4     // select count(id) from user_notification_rules group by user_id order by count desc limit 1
	NRCMMax              = 3     // select count(id) from user_notification_rules group by user_id,contact_method_id order by count desc limit 1
	EPCount              = 371   // select count(id) from escalation_policies
	EPMaxStep            = 4     // select count(id) from escalation_policy_steps group by escalation_policy_id order by count desc limit 1
	EPMaxAssigned        = 5     // select count(id) from escalation_policy_actions group by escalation_policy_step_id order by count desc limit 1
	SvcCount             = 397   // select count(id) from services
	RotationMaxPart      = 50    // select count(id) from rotation_participants group by rotation_id order by count desc limit 1
	ScheduleCount        = 404   // select count(id) from schedules
	AlertClosedCount     = 76909 // select count(id) from alerts where status = 'closed'
	AlertActiveCount     = 2762  // select count(id) from alerts where status = 'triggered' or status = 'active'
	RotationCount        = 529   // select count(id) from rotations
	IntegrationKeyMax    = 11    // select count(id) from integration_keys group by service_id order by count desc limit 1
	ScheduleMaxRules     = 10    // select count(id) from schedule_rules group by schedule_id order by count desc limit 1
	ScheduleMaxOverrides = 24    // select count(*) from user_overrides where end_time > now() group by tgt_schedule_id order by count desc limit 1
	HeartbeatMonitorMax  = 8     // select count(*) from heartbeat_monitors group by service_id order by count desc limit 1
	UserFavMax           = 27    // select count(*) from user_favorites group by user_id order by count desc limit 1)
	SvcLabelMax          = 5     // select count(*) from labels group by tgt_service_id order by count desc limit 1
	UniqueLabelKeys      = 20    // select count(distinct key) from labels
	LabelValueMax        = 13    // select count(distinct value) from labels group by key order by count desc limit 1
	MsgPerAlertMax       = 2
)
