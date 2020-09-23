package smoketest

import (
	"fmt"
	"testing"

	"github.com/target/goalert/smoketest/harness"
)

func TestFixedShifts(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "temp-user"}}, 'foo', ''),
		({{uuid "alt-user"}}, 'foo', ''),
		({{uuid "rule-user"}}, 'bar', '');

	insert into user_contact_methods (id, user_id, name, type, value) 
	values
		({{uuid "temp-cm"}}, {{uuid "temp-user"}}, 'personal', 'SMS', {{phone "temp"}}),
		({{uuid "alt-cm"}}, {{uuid "alt-user"}}, 'personal', 'SMS', {{phone "alt"}}),
		({{uuid "rule-cm"}}, {{uuid "rule-user"}}, 'personal', 'SMS', {{phone "rule"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes) 
	values
		({{uuid "temp-user"}}, {{uuid "temp-cm"}}, 0),
		({{uuid "alt-user"}}, {{uuid "alt-cm"}}, 0),
		({{uuid "rule-user"}}, {{uuid "rule-cm"}}, 0);

	insert into schedules (id, name, time_zone)
	values
		({{uuid "sched"}}, 'sched', 'UTC');
	insert into schedule_rules (id, schedule_id, sunday, monday, tuesday, wednesday, thursday, friday, saturday, start_time, end_time, tgt_user_id)
	values
		({{uuid ""}}, {{uuid "sched"}}, true, true, true, true, true, true, true, '00:00', '00:00', {{uuid "rule-user"}});
	insert into schedule_data (schedule_id, data)
	values ({{uuid "sched"}}, '{"V1":{"TemporarySchedules": [{
		"Start": "0000-08-24T21:03:54Z",
		"End": "9999-08-24T21:03:54Z",
		"Shifts": [{"Start":  "0000-08-24T21:03:54Z", "End": "9998-08-24T21:03:54Z", "UserID": {{uuidJSON "temp-user"}} }]
	}]}}');

	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id) 
	values
		({{uuid "esid"}}, {{uuid "eid"}});
	insert into escalation_policy_actions (escalation_policy_step_id, schedule_id) 
	values 
		({{uuid "esid"}}, {{uuid "sched"}});

	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, summary, dedup_key)
	values
		({{uuid "sid"}}, 'testing', 'auto:1:foo');

`
	h := harness.NewHarness(t, sql, "temp-schedules")
	defer h.Close()

	h.Twilio(t).Device(h.Phone("temp")).ExpectSMS("testing")

	h.GraphQLQueryT(t, fmt.Sprintf(`mutation{setScheduleShifts(input:{
		scheduleID: "%s",
		start: "0000-08-24T21:03:54Z",
		end: "9999-08-24T21:03:54Z",
		shifts: [{start: "0001-08-24T21:03:54Z", end: "9998-08-24T21:03:54Z", userID: "%s"}]
	})}`, h.UUID("sched"), h.UUID("alt-user")))
	h.Trigger()

	h.Escalate(1, 0)

	h.Twilio(t).Device(h.Phone("alt")).ExpectSMS("testing")

	h.GraphQLQueryT(t, fmt.Sprintf(`mutation{resetScheduleShifts(input:{
		scheduleID: "%s",
		start: "0000-08-24T21:03:54Z",
		end: "9999-08-24T21:03:54Z"
	})}`, h.UUID("sched")))
	h.Trigger()

	h.Escalate(1, 0)

	h.Twilio(t).Device(h.Phone("rule")).ExpectSMS("testing")
}
