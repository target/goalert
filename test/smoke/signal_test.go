package smoke

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/test/smoke/harness"
)

// TestSignal tests that signal messages are sent correctly.
func TestSignal(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into users (id, name, email) values
			({{uuid "user"}}, 'bob', 'joe');
		insert into user_contact_methods (id, user_id, name, type, value) values
			({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});
		insert into user_notification_rules (user_id, contact_method_id, delay_minutes) values
			({{uuid "user"}}, {{uuid "cm1"}}, 0);
		insert into escalation_policies (id, name) values
			({{uuid "ep"}}, 'esc policy');
		insert into escalation_policy_steps (id, escalation_policy_id, delay) values
			({{uuid "step"}}, {{uuid "ep"}}, 5);
		insert into escalation_policy_actions (escalation_policy_step_id, user_id) values
			({{uuid "step"}}, {{uuid "user"}});
		insert into services (id, name, escalation_policy_id) values
			({{uuid "svc"}}, 'service', {{uuid "ep"}});
	`

	h := harness.NewHarnessWithFlags(t, sql, "nc-duplicate-table", expflag.FlagSet{expflag.UnivKeys})
	defer h.Close()

	var dest gadb.DestV1
	err := h.App().DB().QueryRowContext(context.Background(), `select dest from user_contact_methods where id = $1`, h.UUID("cm1")).Scan(&dest)
	require.NoError(t, err)

	// validate fields
	assert.Equal(t, "builtin-twilio-sms", dest.Type, "unexpected type")
	assert.Equal(t, h.Phone("1"), dest.Args["phone_number"], "unexpected arg")
}
