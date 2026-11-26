package teams

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
)

func testCtx() context.Context {
	var cfg config.Config
	cfg.General.PublicURL = "http://example.com"
	cfg.Teams.Enable = true
	return cfg.Context(context.Background())
}

func toJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return string(data)
}

func TestAlertCard(t *testing.T) {
	ctx := testCtx()

	card := alertCard(ctx, 123, "Example Summary", "Example Details", "Example Service", "", notification.AlertStateUnacknowledged)
	doc := toJSON(t, card)

	assert.Contains(t, doc, `"Alert #123: Example Summary"`)
	assert.Contains(t, doc, `"Unacknowledged"`)
	assert.Contains(t, doc, `"color":"Attention"`)
	assert.Contains(t, doc, `"Example Details"`)
	assert.Contains(t, doc, `"Example Service"`)
	assert.Contains(t, doc, `"url":"http://example.com/alerts/123"`)
	assert.Contains(t, doc, `"version":"1.2"`)

	card = alertCard(ctx, 123, "Example Summary", "", "", "Acknowledged by Joe", notification.AlertStateAcknowledged)
	doc = toJSON(t, card)
	assert.Contains(t, doc, `"Acknowledged by Joe"`)
	assert.Contains(t, doc, `"color":"Warning"`)

	card = alertCard(ctx, 123, "Example Summary", "", "", "", notification.AlertStateClosed)
	doc = toJSON(t, card)
	assert.Contains(t, doc, `"Closed"`)
	assert.Contains(t, doc, `"color":"Good"`)
}

func TestAlertCard_TruncatesDetails(t *testing.T) {
	long := strings.Repeat("x", maxDetailsLen+100)
	card := alertCard(testCtx(), 1, "s", long, "", "", notification.AlertStateUnacknowledged)
	doc := toJSON(t, card)

	assert.NotContains(t, doc, long)
	assert.Contains(t, doc, strings.Repeat("x", maxDetailsLen)+"…")
}

func TestAlertBundleCard(t *testing.T) {
	card := alertBundleCard(testCtx(), "svc-id", "Example Service", 6)
	doc := toJSON(t, card)

	assert.Contains(t, doc, `"Service 'Example Service' has 6 unacknowledged alerts."`)
	assert.Contains(t, doc, `"url":"http://example.com/services/svc-id/alerts"`)
}

func TestOnCallCard(t *testing.T) {
	card := onCallCard(testCtx(), notification.ScheduleOnCallUsers{
		ScheduleName: "Example Schedule",
		ScheduleURL:  "http://example.com/schedules/sched-id",
		Users: []notification.User{
			{Name: "Alice", URL: "http://example.com/users/alice"},
			{Name: "Bob", URL: "http://example.com/users/bob"},
		},
	})
	doc := toJSON(t, card)

	assert.Contains(t, doc, "[Alice](http://example.com/users/alice)")
	assert.Contains(t, doc, "[Bob](http://example.com/users/bob)")
	assert.Contains(t, doc, `"url":"http://example.com/schedules/sched-id"`)

	card = onCallCard(testCtx(), notification.ScheduleOnCallUsers{
		ScheduleName: "Example Schedule",
		ScheduleURL:  "http://example.com/schedules/sched-id",
	})
	doc = toJSON(t, card)
	assert.Contains(t, doc, "No one is on-call")
}

func TestTestCard(t *testing.T) {
	doc := toJSON(t, testCard(testCtx()))
	assert.Contains(t, doc, "GoAlert Test Notification")
	assert.Contains(t, doc, "This is a test message.")
}
