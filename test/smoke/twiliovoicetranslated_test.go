package smoke

import (
	"testing"
	"time"

	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioVoiceTranslated checks that, when Twilio.VoiceLanguage selects a
// language with a translation, both the spoken alert prompt and the
// acknowledgement confirmation are delivered in that language (here, French)
// rather than the default English.
func TestTwilioVoiceTranslated(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role)
	values
		({{uuid "user"}}, 'bob', 'joe', 'user');
	insert into user_contact_methods (id, user_id, name, type, value)
	values
		({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'VOICE', {{phone "1"}});

	insert into user_notification_rules (user_id, contact_method_id, delay_minutes)
	values
		({{uuid "user"}}, {{uuid "cm1"}}, 0);

	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');
	insert into escalation_policy_steps (id, escalation_policy_id)
	values
		({{uuid "esid"}}, {{uuid "eid"}});
	insert into escalation_policy_actions (escalation_policy_step_id, user_id)
	values
		({{uuid "esid"}}, {{uuid "user"}});

	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (service_id, description)
	values
		({{uuid "sid"}}, 'testing');

`
	h := harness.NewHarness(t, sql, "ids-to-uuids")
	defer h.Close()

	h.SetConfigValue("Twilio.VoiceLanguage", "fr-FR")

	tw := h.Twilio(t)
	d1 := tw.Device(h.Phone("1"))

	// The alert body ("...notification d'alerte...") and the menu prompt
	// ("Pour acquitter...") must both be spoken in French.
	d1.ExpectVoice("notification d'alerte", "pour acquitter").
		ThenPress("4").
		ThenExpect("acquittée")

	h.FastForward(time.Hour)
	// no more messages
}
