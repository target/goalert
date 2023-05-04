package smoke

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/test/smoke/harness"
)

// TestMissingUser tests that notifications go out, even when data is in an odd state.
//
// - escalation policy with no steps
// - escalation policy with steps missing actions
// - policy step with schedule and no users
// - policy step with schedule that starts in the future
func TestMissingUser(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email)
	values
		({{uuid "u1"}}, 'bob', 'joe'),
		({{uuid "u2"}}, 'ben', 'frank');

	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "c1"}}, {{uuid "u1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "c2"}}, {{uuid "u2"}}, 'personal', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "u1"}}, {{uuid "c1"}}, 0),
		({{uuid "u2"}}, {{uuid "c2"}}, 0);

	insert into schedules (id, name, time_zone)
	values
		({{uuid "empty_sched"}}, 'empty', 'America/Chicago'),
		({{uuid "empty_rot_sched"}}, 'empty rot', 'America/Chicago'),
		({{uuid "future_sched"}}, 'future', 'America/Chicago');
	
	insert into rotations (id, schedule_id, name, type, start_time, shift_length)
	values
		({{uuid ""}}, {{uuid "empty_rot_sched"}}, 'def', 'daily', now() - '1 hour'::interval, 1),
		({{uuid "future_rot"}}, {{uuid "future_sched"}}, 'def', 'daily', now() + '1 hour'::interval, 1);
	
	insert into rotation_participants (id, rotation_id, position, user_id)
	values
		({{uuid ""}}, {{uuid "future_rot"}}, 0, {{uuid "u1"}});

	insert into escalation_policies (id, name)
	values
		({{uuid "empty_policy"}}, 'esc policy'),
		({{uuid "empty_step"}}, 'empty step'),
		({{uuid "empty_sched_pol"}}, 'empty sched'),
		({{uuid "empty_rot_pol"}}, 'empty rot'),
		({{uuid "future_sched_pol"}}, 'future'),
		({{uuid "tech.correct"}}, 'woot');

	insert into escalation_policy_steps (id, escalation_policy_id)
	values
		({{uuid ""}}, {{uuid "empty_step"}}),
		({{uuid "empty_sched_step"}}, {{uuid "empty_sched_pol"}}),
		({{uuid "empty_rot_step"}}, {{uuid "empty_rot_pol"}}),
		({{uuid "future_sched_step"}}, {{uuid "future_sched_pol"}}),
		({{uuid "tech.correct_step"}}, {{uuid "tech.correct"}});

	insert into escalation_policy_actions (escalation_policy_step_id, user_id, schedule_id)
	values
		({{uuid "empty_sched_step"}}, null, {{uuid "empty_sched"}}),
		({{uuid "empty_rot_step"}}, null, {{uuid "empty_rot_sched"}}),
		({{uuid "future_sched_step"}}, null, {{uuid "future_sched"}}),
		({{uuid "tech.correct_step"}}, null, {{uuid "empty_sched"}}),
		({{uuid "tech.correct_step"}}, {{uuid "u1"}}, null);

	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "s1"}}, {{uuid "empty_policy"}}, 'service1'),
		({{uuid "s2"}}, {{uuid "empty_step"}}, 'service2'),
		({{uuid "s3"}}, {{uuid "empty_sched_pol"}}, 'service3'),
		({{uuid "s4"}}, {{uuid "future_sched_pol"}}, 'service4'),
		({{uuid "s5"}}, {{uuid "tech.correct"}}, 'service5'),
		({{uuid "s6"}}, {{uuid "empty_rot_pol"}}, 'service6');

	insert into alerts (service_id, description)
	values
		({{uuid "s1"}}, 'emptypol'),
		({{uuid "s2"}}, 'emptystep'),
		({{uuid "s3"}}, 'emptysched'),
		({{uuid "s4"}}, 'futuresched'),
		({{uuid "s5"}}, 'correct'),
		({{uuid "s6"}}, 'emptyrot');

`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	d := h.Twilio(t).Device(h.Phone("1"))
	err := h.EscalateAlertErr(1)
	assert.Error(t, err, "empty policy")
	d.ExpectSMS("correct")

	// Rotations will always have someone active, as long as there are 1 or more participants
}
