package smoketest

import (
	"github.com/target/goalert/smoketest/harness"
	"testing"
	"time"
)

func getLastDSTDate(t *testing.T, n time.Time, loc *time.Location) time.Time {
	const cdt = -18000

	_, offset := n.Zone()

	if offset == cdt {
		for offset == cdt {
			n = n.AddDate(0, -1, 0)
			_, offset = n.Zone()
		}
	} else {
		for offset != cdt {
			n = n.AddDate(0, -1, 0)
			_, offset = n.Zone()
		}

	}

	return n
}

// TestRotation_DST checks that schedules handle DST boundaries properly
func TestRotation_DST(t *testing.T) {
	t.Parallel()
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatalf("could not load 'America/Chicago' tzdata: %v", err)
	}

	// for this test, we make a daily rotation at the current time (- 1 minute) across the closest DST boundary
	// then make sure the rotation flips after a minute.
	// make it easy

	n := time.Now().In(loc)
	start := getLastDSTDate(t, n, loc).Add(5 * time.Minute)
	startStr := start.Format(time.RFC3339)
	sql := `
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
		({{uuid "rot1"}}, 'default rotation', 'daily', '` + startStr + `',1, 'America/Chicago');

	insert into rotation_participants (rotation_id, user_id, position)
	values
		({{uuid "rot1"}}, {{uuid "uid1"}}, 0),
		({{uuid "rot1"}}, {{uuid "uid2"}}, 1);

	insert into escalation_policy_actions (escalation_policy_step_id, rotation_id)
	values
		({{uuid "esid"}}, {{uuid "rot1"}});

	insert into services (id, escalation_policy_id, name) values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, description) values
		({{uuid "sid"}}, 'testing');

	`
	h := harness.NewHarness(t, sql, "ev3-rotation-state")
	defer h.Close()

	sid := h.UUID("sid")
	uid1 := h.UUID("uid1")
	uid2 := h.UUID("uid2")

	h.WaitAndAssertOnCallUsers(sid, uid1)

	h.FastForward(10 * time.Minute)

	h.WaitAndAssertOnCallUsers(sid, uid2)
}
