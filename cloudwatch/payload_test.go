package cloudwatch

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/alert"
)

const (
	testTopicARN  = "arn:aws:sns:us-west-2:123456789012:PagerDuty-Data"
	testAlarmName = "[us-west-2] Too Many Write Errors"
	testRunbook   = "Runbook: https://runbook.example.com/dp/ats-write-errors"
	testServiceID = "3c1a1a44-8e7a-4d1b-9f4a-2b0e5c6d7f80"
)

// goldenAlarm is the CloudWatch alarm body used by the golden-path cases.
const goldenAlarm = `{
	"AlarmName": "[us-west-2] Too Many Write Errors",
	"AlarmDescription": "Runbook: https://runbook.example.com/dp/ats-write-errors",
	"AWSAccountId": "123456789012",
	"NewStateValue": "ALARM",
	"OldStateValue": "OK",
	"NewStateReason": "Threshold Crossed: 1 datapoint was greater than the threshold (1.0).",
	"StateChangeTime": "2026-07-30T00:00:00.000+0000",
	"AlarmArn": "arn:aws:cloudwatch:us-west-2:123456789012:alarm:x",
	"Region": "US West (Oregon)",
	"Trigger": {"Namespace": "Eightfold/DP", "MetricName": "WriteErrors"}
}`

const goldenDetails = `State: OK -> ALARM
Reason: Threshold Crossed: 1 datapoint was greater than the threshold (1.0).
Changed: 2026-07-30T00:00:00.000+0000
Region: us-west-2
Account: 123456789012
Metric: Eightfold/DP/WriteErrors
Topic: PagerDuty-Data
Alarm ARN: arn:aws:cloudwatch:us-west-2:123456789012:alarm:x

Runbook: https://runbook.example.com/dp/ats-write-errors`

func notification(message string) envelope {
	return envelope{Type: typeNotification, TopicARN: testTopicARN, Message: message}
}

func TestSplitTopicARN(t *testing.T) {
	tests := []struct {
		name, arn, region, topic string
	}{
		{name: "full arn", arn: testTopicARN, region: "us-west-2", topic: "PagerDuty-Data"},
		{name: "empty", arn: "", region: "", topic: ""},
		{name: "no colons", arn: "topic", region: "", topic: "topic"},
		{name: "short arn does not panic", arn: "arn:aws:sns", region: "", topic: "sns"},
		{name: "exactly four segments", arn: "arn:aws:sns:us-west-2", region: "us-west-2", topic: "us-west-2"},
		{name: "extra colons take last", arn: "arn:aws:sns:us-west-2:123:my:topic", region: "us-west-2", topic: "topic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, topic := splitTopicARN(tt.arn)
			assert.Equal(t, tt.region, region, "region")
			assert.Equal(t, tt.topic, topic, "topic")
		})
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{name: "empty", in: "", want: ""},
		{name: "single line", in: "a", want: "a"},
		{name: "lf", in: "a\nb", want: "a"},
		{name: "crlf", in: "a\r\nb", want: "a"},
		{name: "leading and trailing blank lines", in: "\n\na\n", want: "a"},
		{name: "whitespace only", in: "   \n  ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, firstLine(tt.in))
		})
	}
}

func TestBuildAlert_CloudWatch(t *testing.T) {
	goldenDedup := sha256Hex(testAlarmName)

	t.Run("golden alarm", func(t *testing.T) {
		a, meta, ok := buildAlert(notification(goldenAlarm))
		require.True(t, ok)

		assert.Equal(t, testAlarmName, a.Summary)
		assert.Equal(t, goldenDetails, a.Details)
		assert.Equal(t, alert.StatusTriggered, a.Status)
		assert.Equal(t, alert.SourceCloudwatch, a.Source)
		require.NotNil(t, a.Dedup)
		assert.Equal(t, goldenDedup, a.Dedup.Payload)
		assert.Equal(t, map[string]string{
			"topic":       "PagerDuty-Data",
			"region":      "us-west-2",
			"state":       "ALARM",
			"aws_account": "123456789012",
			"namespace":   "Eightfold/DP",
			"metric":      "WriteErrors",
			"alarm_arn":   "arn:aws:cloudwatch:us-west-2:123456789012:alarm:x",
		}, meta)
	})

	// The dedup key must be identical to the ALARM case or the close never
	// matches the open alert.
	t.Run("OK closes with the same dedup", func(t *testing.T) {
		body := strings.Replace(goldenAlarm, `"NewStateValue": "ALARM"`, `"NewStateValue": "OK"`, 1)
		a, meta, ok := buildAlert(notification(body))
		require.True(t, ok)

		assert.Equal(t, alert.StatusClosed, a.Status)
		require.NotNil(t, a.Dedup)
		assert.Equal(t, goldenDedup, a.Dedup.Payload)
		assert.Equal(t, "OK", meta["state"])
	})

	t.Run("INSUFFICIENT_DATA creates nothing", func(t *testing.T) {
		body := strings.Replace(goldenAlarm, `"NewStateValue": "ALARM"`, `"NewStateValue": "INSUFFICIENT_DATA"`, 1)
		_, meta, ok := buildAlert(notification(body))
		assert.False(t, ok)
		assert.Nil(t, meta)
	})

	t.Run("insufficient_data is case insensitive", func(t *testing.T) {
		body := strings.Replace(goldenAlarm, `"NewStateValue": "ALARM"`, `"NewStateValue": "insufficient_data"`, 1)
		_, _, ok := buildAlert(notification(body))
		assert.False(t, ok)
	})

	// The reference implementation compares state == "OK" exactly, so a
	// lowercase "ok" triggers rather than closes. CloudWatch only ever sends
	// uppercase; this pins the asymmetry with INSUFFICIENT_DATA above.
	t.Run("lowercase ok triggers", func(t *testing.T) {
		body := strings.Replace(goldenAlarm, `"NewStateValue": "ALARM"`, `"NewStateValue": "ok"`, 1)
		a, _, ok := buildAlert(notification(body))
		require.True(t, ok)
		assert.Equal(t, alert.StatusTriggered, a.Status)
	})

	t.Run("no description means no trailing blank line", func(t *testing.T) {
		body := strings.Replace(goldenAlarm, `"AlarmDescription": "`+testRunbook+`",`, "", 1)
		a, _, ok := buildAlert(notification(body))
		require.True(t, ok)

		assert.True(t, strings.HasSuffix(a.Details, "Alarm ARN: arn:aws:cloudwatch:us-west-2:123456789012:alarm:x"), a.Details)
		assert.NotContains(t, a.Details, testRunbook)
	})

	t.Run("null description", func(t *testing.T) {
		body := strings.Replace(goldenAlarm, `"AlarmDescription": "`+testRunbook+`"`, `"AlarmDescription": null`, 1)
		a, _, ok := buildAlert(notification(body))
		require.True(t, ok)
		assert.NotContains(t, a.Details, testRunbook)
	})

	t.Run("minimal fields", func(t *testing.T) {
		e := envelope{Type: typeNotification, Message: `{"AlarmName":"x","NewStateValue":"ALARM"}`}
		a, meta, ok := buildAlert(e)
		require.True(t, ok)

		// The State line is unconditional even with an empty old state.
		assert.Equal(t, "State:  -> ALARM", a.Details)
		assert.Equal(t, map[string]string{"state": "ALARM"}, meta)
	})

	t.Run("metric line requires a metric name", func(t *testing.T) {
		body := strings.Replace(goldenAlarm, `"MetricName": "WriteErrors"`, `"MetricName": ""`, 1)
		a, meta, ok := buildAlert(notification(body))
		require.True(t, ok)

		assert.NotContains(t, a.Details, "Metric:")
		assert.Equal(t, "Eightfold/DP", meta["namespace"])
		assert.NotContains(t, meta, "metric")
	})

	t.Run("metric name without namespace keeps leading slash", func(t *testing.T) {
		body := strings.Replace(goldenAlarm, `"Namespace": "Eightfold/DP"`, `"Namespace": ""`, 1)
		a, _, ok := buildAlert(notification(body))
		require.True(t, ok)
		assert.Contains(t, a.Details, "Metric: /WriteErrors")
	})

	// Summary is truncated but the dedup hashes the full name, so truncation can
	// never collapse two distinct alarms onto one alert.
	t.Run("summary truncated but dedup is not", func(t *testing.T) {
		long := strings.Repeat("x", 1200)
		a, _, ok := buildAlert(notification(`{"AlarmName":"` + long + `","NewStateValue":"ALARM"}`))
		require.True(t, ok)

		assert.Len(t, []rune(a.Summary), alert.MaxSummaryLength)
		assert.True(t, strings.HasSuffix(a.Summary, "…"))
		require.NotNil(t, a.Dedup)
		assert.Equal(t, sha256Hex(long), a.Dedup.Payload)
	})

	// The whole point of capping NewStateReason: it is the one unbounded field, so
	// without the cap it would crowd the runbook URL past the details limit.
	// Capping it keeps the whole body comfortably under the limit instead.
	t.Run("huge reason still keeps the runbook", func(t *testing.T) {
		body := strings.Replace(goldenAlarm,
			`"NewStateReason": "Threshold Crossed: 1 datapoint was greater than the threshold (1.0)."`,
			`"NewStateReason": "`+strings.Repeat("r", 10000)+`"`, 1)
		a, _, ok := buildAlert(notification(body))
		require.True(t, ok)

		assert.LessOrEqual(t, len([]rune(a.Details)), alert.MaxDetailsLength)
		assert.NotContains(t, a.Details, "…", "details should fit without truncation")
		assert.Contains(t, a.Details, testRunbook)
		assert.Contains(t, a.Details, "Reason: "+strings.Repeat("r", maxReasonLen)+"\n")
		assert.NotContains(t, a.Details, strings.Repeat("r", maxReasonLen+1))
	})

	t.Run("region comes from the arn not the display name", func(t *testing.T) {
		a, meta, ok := buildAlert(notification(goldenAlarm))
		require.True(t, ok)

		assert.Contains(t, a.Details, "Region: us-west-2")
		assert.Equal(t, "us-west-2", meta["region"])
		assert.NotContains(t, a.Details, "Oregon")
		for k, v := range meta {
			assert.NotContains(t, v, "Oregon", "meta[%s]", k)
		}
	})

	t.Run("malformed arn", func(t *testing.T) {
		e := envelope{Type: typeNotification, TopicARN: "arn:aws:sns", Message: goldenAlarm}
		a, meta, ok := buildAlert(e)
		require.True(t, ok)

		assert.NotContains(t, a.Details, "Region:")
		assert.Contains(t, a.Details, "Topic: sns")
		assert.Equal(t, "sns", meta["topic"])
		assert.NotContains(t, meta, "region")
	})

	t.Run("empty arn", func(t *testing.T) {
		e := envelope{Type: typeNotification, Message: goldenAlarm}
		a, meta, ok := buildAlert(e)
		require.True(t, ok)

		assert.NotContains(t, a.Details, "Region:")
		assert.NotContains(t, a.Details, "Topic:")
		assert.NotContains(t, meta, "topic")
		assert.NotContains(t, meta, "region")
		assert.Equal(t, testAlarmName, a.Summary)
	})

	// A blank-but-present AlarmName would otherwise sanitize to an empty summary,
	// which alert.Normalize accepts, producing an unactionable alert.
	t.Run("blank alarm name gets the fallback", func(t *testing.T) {
		a, _, ok := buildAlert(notification(`{"AlarmName":"   ","NewStateValue":"ALARM"}`))
		require.True(t, ok)

		assert.Equal(t, "unnamed alarm on PagerDuty-Data", a.Summary)
		require.NotNil(t, a.Dedup)
		assert.Equal(t, sha256Hex("unnamed alarm on PagerDuty-Data"), a.Dedup.Payload)
	})

	// encoding/json fills what it can, so one wrong-typed field must not cost us
	// the CloudWatch branch and its dedup contract.
	t.Run("wrong typed field still maps as an alarm", func(t *testing.T) {
		a, meta, ok := buildAlert(notification(`{"AlarmName":"x","AWSAccountId":12345,"NewStateValue":"ALARM"}`))
		require.True(t, ok)

		assert.Equal(t, "x", a.Summary)
		assert.NotContains(t, a.Details, "Account:")
		assert.NotContains(t, meta, "aws_account")
	})

	t.Run("non string alarm name falls through to raw", func(t *testing.T) {
		_, meta, ok := buildAlert(notification(`{"AlarmName":123,"x":"y"}`))
		require.True(t, ok)
		assert.Equal(t, "sns-raw", meta["source"])
	})
}

func TestBuildAlert_Raw(t *testing.T) {
	t.Run("subject wins", func(t *testing.T) {
		e := notification("line one\nline two")
		e.Subject = strPtr("Backup failed")
		a, meta, ok := buildAlert(e)
		require.True(t, ok)

		assert.Equal(t, "Backup failed", a.Summary)
		assert.Equal(t, "line one\nline two", a.Details)
		assert.Equal(t, alert.StatusTriggered, a.Status)
		require.NotNil(t, a.Dedup)
		assert.Equal(t, sha256Hex("PagerDuty-Data|Backup failed"), a.Dedup.Payload)
		assert.Equal(t, map[string]string{
			"topic":  "PagerDuty-Data",
			"region": "us-west-2",
			"source": "sns-raw",
		}, meta)
	})

	t.Run("nil subject uses first line", func(t *testing.T) {
		a, _, ok := buildAlert(notification("first line\nsecond"))
		require.True(t, ok)
		assert.Equal(t, "first line", a.Summary)
	})

	t.Run("empty subject uses first line", func(t *testing.T) {
		e := notification("hello")
		e.Subject = strPtr("")
		a, _, ok := buildAlert(e)
		require.True(t, ok)
		assert.Equal(t, "hello", a.Summary)
	})

	// Python's `subject or ...` treats "   " as truthy, which would sanitize to
	// an empty summary.
	t.Run("whitespace subject falls through", func(t *testing.T) {
		e := notification("hello")
		e.Subject = strPtr("   ")
		a, _, ok := buildAlert(e)
		require.True(t, ok)
		assert.Equal(t, "hello", a.Summary)
	})

	t.Run("no subject and no message", func(t *testing.T) {
		a, _, ok := buildAlert(notification(""))
		require.True(t, ok)
		assert.Equal(t, "SNS notification on PagerDuty-Data", a.Summary)
		assert.Empty(t, a.Details)
	})

	t.Run("no subject no message no arn is still non empty", func(t *testing.T) {
		a, _, ok := buildAlert(envelope{Type: typeNotification})
		require.True(t, ok)
		// Trailing space is trimmed by SanitizeText.
		assert.Equal(t, "SNS notification on", a.Summary)
	})

	t.Run("plain text body", func(t *testing.T) {
		a, _, ok := buildAlert(notification("just some text"))
		require.True(t, ok)
		assert.Equal(t, "just some text", a.Summary)
	})

	t.Run("json array body", func(t *testing.T) {
		a, _, ok := buildAlert(notification("[1,2,3]"))
		require.True(t, ok)
		assert.Equal(t, "[1,2,3]", a.Summary)
	})

	t.Run("json object without alarm name", func(t *testing.T) {
		a, _, ok := buildAlert(notification(`{"foo":"bar"}`))
		require.True(t, ok)
		assert.Equal(t, `{"foo":"bar"}`, a.Summary)
	})

	// This is why details are sanitized rather than raw-sliced: alert.Normalize
	// rejects text that begins with a space or holds non-printables, which would
	// otherwise 400 forever while SNS retried.
	t.Run("leading newlines sanitized", func(t *testing.T) {
		a, _, ok := buildAlert(notification("\n\n  hello  \n"))
		require.True(t, ok)
		assert.Equal(t, "hello", a.Details)
		assert.Equal(t, "hello", a.Summary)
	})

	t.Run("control chars stripped", func(t *testing.T) {
		a, _, ok := buildAlert(notification("bad\x00char"))
		require.True(t, ok)
		assert.Equal(t, "badchar", a.Details)
	})

	t.Run("over limit details truncated", func(t *testing.T) {
		a, _, ok := buildAlert(notification(strings.Repeat("d", 300000)))
		require.True(t, ok)
		assert.Len(t, []rune(a.Details), alert.MaxDetailsLength)
		assert.True(t, strings.HasSuffix(a.Details, "…"))
	})

	t.Run("same subject and topic dedup identically", func(t *testing.T) {
		a := notification("body one")
		a.Subject = strPtr("same")
		b := notification("body two")
		b.Subject = strPtr("same")

		one, _, ok := buildAlert(a)
		require.True(t, ok)
		two, _, ok := buildAlert(b)
		require.True(t, ok)

		require.NotNil(t, one.Dedup)
		require.NotNil(t, two.Dedup)
		assert.Equal(t, one.Dedup.Payload, two.Dedup.Payload)
	})
}

// Invariants that must hold for every mapped notification, checked across all
// branches at once.
func TestBuildAlert_Invariants(t *testing.T) {
	longName := strings.Repeat("x", 1200)
	cases := map[string]envelope{
		"golden":               notification(goldenAlarm),
		"ok":                   notification(strings.Replace(goldenAlarm, `"NewStateValue": "ALARM"`, `"NewStateValue": "OK"`, 1)),
		"minimal alarm":        notification(`{"AlarmName":"x","NewStateValue":"ALARM"}`),
		"blank alarm name":     notification(`{"AlarmName":"   ","NewStateValue":"ALARM"}`),
		"long alarm name":      notification(`{"AlarmName":"` + longName + `","NewStateValue":"ALARM"}`),
		"malformed arn":        {Type: typeNotification, TopicARN: "arn:aws:sns", Message: goldenAlarm},
		"empty arn":            {Type: typeNotification, Message: goldenAlarm},
		"raw plain":            notification("just some text"),
		"raw empty":            notification(""),
		"raw nothing at all":   {Type: typeNotification},
		"raw control chars":    notification("bad\x00char"),
		"raw leading newlines": notification("\n\n  hello  \n"),
		"raw huge":             notification(strings.Repeat("d", 300000)),
	}

	for name, e := range cases {
		t.Run(name, func(t *testing.T) {
			a, meta, ok := buildAlert(e)
			require.True(t, ok)

			// Never fall back to the auto content hash: it changes on every state
			// transition, which would break idempotency and the OK close.
			require.NotNil(t, a.Dedup, "dedup must never be nil")
			assert.Equal(t, alert.DedupTypeUser, a.Dedup.Type)
			assert.Equal(t, 1, a.Dedup.Version)
			assert.Len(t, a.Dedup.Payload, 64, "dedup payload should be a hex sha256")
			assert.Same(t, a.Dedup, a.DedupKey())

			assert.Equal(t, alert.SourceCloudwatch, a.Source)
			assert.NotEmpty(t, a.Summary, "summary must never be empty")

			// Proves the mapper can never produce a client error, and catches a
			// missing SourceCloudwatch entry in alert.Normalize's OneOf.
			a.ServiceID = testServiceID
			_, err := a.Normalize()
			assert.NoError(t, err)

			total := 0
			for k, v := range meta {
				assert.NotEmpty(t, v, "meta[%s] should have been dropped", k)
				assert.LessOrEqual(t, len([]rune(v)), maxMetaValueLen, "meta[%s]", k)
				total += len(k) + len(v)
			}
			assert.Less(t, total, 32*1024, "metadata must stay inside the store cap")
		})
	}
}

// alert.Normalize collapses newlines in the summary and does a single pass of
// double-space replacement, so three spaces become two rather than one.
func TestNormalizeSummaryQuirks(t *testing.T) {
	t.Run("three spaces collapse to two", func(t *testing.T) {
		a, _, ok := buildAlert(notification(`{"AlarmName":"[us-west-2]   Too Many Errors","NewStateValue":"ALARM"}`))
		require.True(t, ok)
		assert.Equal(t, "[us-west-2]   Too Many Errors", a.Summary)

		a.ServiceID = testServiceID
		n, err := a.Normalize()
		require.NoError(t, err)
		assert.Equal(t, "[us-west-2]  Too Many Errors", n.Summary)
	})

	t.Run("newlines become a single space", func(t *testing.T) {
		a, _, ok := buildAlert(notification(`{"AlarmName":"a\n\n\nb","NewStateValue":"ALARM"}`))
		require.True(t, ok)
		assert.Equal(t, "a\n\nb", a.Summary)

		a.ServiceID = testServiceID
		n, err := a.Normalize()
		require.NoError(t, err)
		assert.Equal(t, "a b", n.Summary)
	})
}
