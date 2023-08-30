package smoke

import (
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
)

func TestRotation_Monthly(t *testing.T) {
	t.Parallel()

	const sql = `
	set timezone = 'America/Chicago';

	insert into users (id, name, email)
	values
		({{uuid "uid1"}}, 'bob', 'joe'),
		({{uuid "uid2"}}, 'ben', 'frank');

	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "uid1"}}, 'personal', 'SMS', {{phone "1"}}),
		({{uuid "cm2"}}, {{uuid "uid2"}}, 'personal', 'SMS', {{phone "2"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "uid1"}}, {{uuid "cm1"}}, 0),
		({{uuid "uid2"}}, {{uuid "cm2"}}, 0);

	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id)
	values
		({{uuid "esid"}}, {{uuid "eid"}});

	insert into rotations (id, name, type, start_time, shift_length, time_zone)
	values
		({{uuid "rot1"}}, 'default rotation', 'monthly', now(), 1, 'America/Chicago');

	insert into rotation_participants (rotation_id, user_id, position)
	values
		({{uuid "rot1"}}, {{uuid "uid1"}}, 0),
		({{uuid "rot1"}}, {{uuid "uid2"}}, 1);

	insert into escalation_policy_actions (escalation_policy_step_id, rotation_id)
	values
		({{uuid "esid"}}, {{uuid "rot1"}});

	insert into services (id, escalation_policy_id, name) values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	`

	h := harness.NewHarness(t, sql, "add-monthly-rotation")
	defer h.Close()

	sid := h.UUID("sid")
	uid1 := h.UUID("uid1")
	uid2 := h.UUID("uid2")

	h.WaitAndAssertOnCallUsers(sid, uid1)

	h.FastForwardExtended(32 * 24 * time.Hour)

	h.WaitAndAssertOnCallUsers(sid, uid2)
}
