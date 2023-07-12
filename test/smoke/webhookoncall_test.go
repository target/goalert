package smoke

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/test/smoke/harness"
)

type POSTDataOnCallUser struct {
	ID   string
	Name string
	URL  string
}

type POSTDataOnCallNotification struct {
	AppName      string
	Type         string
	Users        []POSTDataOnCallUser
	ScheduleID   string
	ScheduleName string
	ScheduleURL  string
}

// TestWebhookOnCallNotification tests that the configured rule sends the intended notification to the webhook.
func TestWebhookOnCallNotification(t *testing.T) {
	t.Parallel()

	ch := make(chan POSTDataOnCallNotification, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var alert POSTDataOnCallNotification

		data, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err) {
			return
		}

		err = json.Unmarshal(data, &alert)
		if !assert.NoError(t, err) {
			return
		}

		ch <- alert
	}))

	defer ts.Close()

	sql := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	
	insert into schedules (id, name, time_zone) 
	values
		({{uuid "sid"}}, 'testschedule', 'UTC');

	insert into schedule_rules (id, schedule_id, sunday, monday, tuesday, wednesday, thursday, friday, saturday, start_time, end_time, tgt_user_id)
	values
		({{uuid "ruleID"}}, {{uuid "sid"}}, true, true, true, true, true, true, true, '00:00:00', '00:00:00', {{uuid "user"}});
	
	insert into notification_channels (id, type, name, value)
	values
		({{uuid "webhook"}}, 'WEBHOOK', 'url', '` + ts.URL + `');

	insert into schedule_data (schedule_id, data)
	values
		({{uuid "sid"}}, '{"V1":{"OnCallNotificationRules": [{"ChannelID": {{uuidJSON "webhook"}}, "Time": "00:00" }]}}');
	`

	h := harness.NewHarnessWithFlags(t, sql, "webhook-notification-channel-type", expflag.FlagSet{expflag.ChanWebhook})
	defer h.Close()

	h.Trigger()

	h.FastForward(24 * time.Hour)

	alert := <-ch
	assert.Equal(t, alert.Users[0].Name, "bob")
	assert.Equal(t, alert.ScheduleName, "testschedule")
}
