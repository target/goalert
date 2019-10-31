package twilio

import (
	"strconv"
	"strings"
	"testing"
)

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

func TestAlertSMS_Render(t *testing.T) {
	check := func(name string, a alertSMS, exp string) {
		t.Run(name, func(t *testing.T) {
			res, err := a.Render()
			if len(res) > 160 {
				t.Errorf("message exceeded 160 characters")
			} else {
				t.Log("Length", len(res))
			}
			if err != nil && exp != "" {
				t.Fatalf("got err %v; want nil", err)
			} else if err == nil && exp == "" {
				t.Log(res)
				t.Fatal("got nil; want err")
			}

			if res != exp {
				t.Errorf("got %s; want %s", strconv.Quote(res), strconv.Quote(exp))
			}
		})
	}

	check("normal",
		alertSMS{
			ID:   123,
			Code: 1,
			Link: "https://example.com/alerts/123",
			Body: "Testing",
		},
		`Alert #123: Testing

https://example.com/alerts/123

Reply '1a' to ack, '1c' to close.`,
	)

	check("no-reply",
		alertSMS{
			ID:   123,
			Link: "https://example.com/alerts/123",
			Body: "Testing",
		},
		`Alert #123: Testing

https://example.com/alerts/123`,
	)

	check("no-Link",
		alertSMS{
			ID:   123,
			Code: 1,
			Body: "Testing",
		},
		`Alert #123: Testing

Reply '1a' to ack, '1c' to close.`,
	)

	check("no-reply-Link",
		alertSMS{
			ID:   123,
			Body: "Testing",
		},
		`Alert #123: Testing`,
	)

	check("truncate",
		alertSMS{
			ID:   123,
			Code: 1,
			Link: "https://example.com/alerts/123",
			Body: "Testing with a really really obnoxiously long message that will be need to be truncated at some point.",
		},
		`Alert #123: Testing with a really really obnoxiously long message that will be need to be tru

https://example.com/alerts/123

Reply '1a' to ack, '1c' to close.`,
	)

	check("truncate-long-id",
		alertSMS{
			ID:   123456789,
			Code: 1,
			Link: "https://example.com/alerts/123",
			Body: "Testing with a really really obnoxiously long message that will be need to be truncated at some point.",
		},
		`Alert #123456789: Testing with a really really obnoxiously long message that will be need to

https://example.com/alerts/123

Reply '1a' to ack, '1c' to close.`,
	)

	check("message-too-long",
		// can't fit body
		alertSMS{
			ID:   123456789,
			Code: 123456789,
			Link: "https://example.com/alerts/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff/123ff",
			Body: "Testing with a really really obnoxiously long message that will be need to be truncated at some point.",
		},
		"",
	)

	check("alert-bundle-one",
		alertSMS{
			Count: 1,
			Body:  "My Service",
			Code:  100,
			Link:  "https://example.com/services/321-654/alerts",
		},
		`Svc 'My Service': 1 unacked alert

https://example.com/services/321-654/alerts

Reply '100aa' to ack all, '100cc' to close all.`,
	)

	check("alert-bundle",
		alertSMS{
			Count: 5,
			Body:  "My Service",
			Code:  100,
			Link:  "https://example.com/services/321-654/alerts",
		},
		`Svc 'My Service': 5 unacked alerts

https://example.com/services/321-654/alerts

Reply '100aa' to ack all, '100cc' to close all.`,
	)

	check("status-bundle-one",
		// can't fit body
		alertSMS{
			ID:    123,
			Count: 1,
			Body:  "Some log entry",
		},
		`Alert #123: Some log entry

1 other alert has been updated.`,
	)

	check("status-bundle",
		// can't fit body
		alertSMS{
			ID:    123,
			Count: 2,
			Body:  "Some log entry",
		},
		`Alert #123: Some log entry

2 other alerts have been updated.`,
	)

}
