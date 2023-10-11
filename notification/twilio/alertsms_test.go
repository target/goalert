package twilio

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func resultCheck(t *testing.T, expected string, res string, err error) {
	t.Helper()
	t.Log("Length", len(res))
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestMapGSM(t *testing.T) {
	check := func(input, exp string) {
		t.Run("", func(t *testing.T) {
			res := strings.Map(mapGSM, input)
			if res != exp {
				t.Errorf("got %s; want %s", strconv.Quote(res), strconv.Quote(exp))
			}
		})
	}

	check("foo\\bar", "foo?bar")
	check("foo\bar", "fooar")
	check("foo\nbar", "foo bar")
	check("foo\t bar@/ok:asdf", "foo  bar@/ok:asdf")
	check("[Testing] {alert_message: `okay`}", "(Testing) (alert-message: 'okay')")
}

func TestSMS_RenderAlert(t *testing.T) {
	check := func(name string, a notification.Alert, link string, code int, exp string) {
		t.Run(name, func(t *testing.T) {
			res, err := renderAlertMessage("TestApp", a, link, code)
			resultCheck(t, exp, res, err)
		})
	}

	check("normal",
		notification.Alert{
			AlertID: 123,
			Summary: "Testing",
		},
		"https://example.com/alerts/123",
		1,
		`TestApp: Alert #123: Testing

https://example.com/alerts/123

Reply '1a' to ack, '1e' to escalate, '1c' to close.`,
	)

	check("no-reply-code",
		notification.Alert{
			AlertID: 123,
			Summary: "Testing",
		},
		"https://example.com/alerts/123",
		0,
		`TestApp: Alert #123: Testing

https://example.com/alerts/123`,
	)

	check("no-link",
		notification.Alert{
			AlertID: 123,
			Summary: "Testing",
		},
		"",
		1,
		`TestApp: Alert #123: Testing

Reply '1a' to ack, '1e' to escalate, '1c' to close.`,
	)

	check("no-link-or-reply-code",
		notification.Alert{
			AlertID: 123,
			Summary: "Testing",
		},
		"",
		0,
		`TestApp: Alert #123: Testing`,
	)

	check("truncate",
		notification.Alert{
			AlertID: 123,
			Summary: "Testing with a really really obnoxiously long message that will be need to be truncated at some point.",
		},
		"https://example.com/alerts/123",
		1,
		`TestApp: Alert #123: Testing with a really really obnoxiously long message

https://example.com/alerts/123

Reply '1a' to ack, '1e' to escalate, '1c' to close.`,
	)

	check("truncate-long-id",
		notification.Alert{
			AlertID: 123456789,
			Summary: "Testing with a really really obnoxiously long message that will be need to be truncated at some point.",
		},
		"https://example.com/alerts/123",
		1,
		`TestApp: Alert #123456789: Testing with a really really obnoxiously long me

https://example.com/alerts/123

Reply '1a' to ack, '1e' to escalate, '1c' to close.`,
	)

	check("message-too-long",
		// can't fit summary
		notification.Alert{
			AlertID: 123456789,
			Summary: "Testing with a really really obnoxiously long message that will be need to be truncated at some point.",
		},
		"https://example.com/alerts/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff",
		123456789,
		`TestApp: Alert #123456789: 

https://example.com/alerts/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff

Reply '123456789a' to ack, '123456789e' to escalate, '123456789c' to close.`,
	)
}

func TestSMS_RenderAlertBundle(t *testing.T) {
	check := func(name string, a notification.AlertBundle, link string, code int, exp string) {
		t.Run(name, func(t *testing.T) {
			res, err := renderAlertBundleMessage("TestApp", a, link, code)
			resultCheck(t, exp, res, err)
		})
	}

	check("alert-bundle-one",
		notification.AlertBundle{
			Count:       1,
			ServiceName: "My Service",
		},
		"https://example.com/services/321-654/alerts",
		100,
		`TestApp: Svc 'My Service': 1 unacked alert

	https://example.com/services/321-654/alerts

	Reply '100aa' to ack all, '100cc' to close all.`,
	)

	check("alert-bundle",
		notification.AlertBundle{
			Count:       5,
			ServiceName: "My Service",
		},
		"https://example.com/services/321-654/alerts",
		100,
		`TestApp: Svc 'My Service': 5 unacked alerts

	https://example.com/services/321-654/alerts

	Reply '100aa' to ack all, '100cc' to close all.`,
	)
}

func TestSMS_RenderAlertStatus(t *testing.T) {
	check := func(name string, a notification.AlertStatus, exp string) {
		t.Run(name, func(t *testing.T) {
			res, err := renderAlertStatusMessage("TestApp", a)
			resultCheck(t, exp, res, err)
		})
	}

	check("alert-status",
		notification.AlertStatus{
			AlertID:  123,
			Summary:  "Testing",
			LogEntry: "Some log entry",
		},
		`TestApp: Alert #123: Testing

	Some log entry`,
	)
}
