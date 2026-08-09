package cloudwatch

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/validation/validate"
)

const (
	// maxReasonLen bounds NewStateReason, the one unbounded field in practice.
	// Without it a verbose reason pushes AlarmDescription -- which carries the
	// runbook URL -- past MaxDetailsLength and it gets truncated away.
	maxReasonLen = 2048

	// maxMetaValueLen keeps metadata well inside alert.ValidateMetadata's total
	// cap. Values come straight from the payload, and exceeding the cap is a
	// client error, which SNS would retry forever.
	maxMetaValueLen = 1024
)

// cloudWatchAlarm is the subset of a CloudWatch alarm notification we map.
type cloudWatchAlarm struct {
	AlarmName        string `json:"AlarmName"`
	AlarmDescription string `json:"AlarmDescription"`
	AWSAccountID     string `json:"AWSAccountId"`
	NewStateValue    string `json:"NewStateValue"`
	OldStateValue    string `json:"OldStateValue"`
	NewStateReason   string `json:"NewStateReason"`
	StateChangeTime  string `json:"StateChangeTime"`
	AlarmARN         string `json:"AlarmArn"`

	// Region is CloudWatch's display name, e.g. "US West (Oregon)". Deliberately
	// unused: region always comes from the topic ARN.
	Region string `json:"Region"`

	Trigger struct {
		Namespace  string `json:"Namespace"`
		MetricName string `json:"MetricName"`
	} `json:"Trigger"`
}

// buildAlert maps a verified SNS Notification onto an alert.
//
// The returned alert has ServiceID unset; the caller fills it in from the
// integration key. Source is always alert.SourceCloudwatch and Dedup is always
// non-nil -- a nil Dedup would silently fall back to a content hash that changes
// on every state transition, breaking both idempotency and the OK close.
//
// ok is false when the notification is intentionally ignored and nothing at all
// should be created (today: CloudWatch INSUFFICIENT_DATA). There is no error
// return: any body we cannot read as a CloudWatch alarm is mapped as a raw
// notification instead of failing, so this can never produce a 4xx.
func buildAlert(e envelope) (a alert.Alert, meta map[string]string, ok bool) {
	region, topic := splitTopicARN(e.TopicARN)

	var alarm cloudWatchAlarm
	err := json.Unmarshal([]byte(e.Message), &alarm)

	// Tolerate a type mismatch on some other field: encoding/json still fills
	// everything it could decode, so a numeric AWSAccountId maps as an alarm
	// minus that one value rather than losing the dedup contract entirely.
	// Note the branch test is a bare non-empty check, matching the reference's
	// truthiness gate: a whitespace-only AlarmName is still a CloudWatch alarm,
	// and buildAlarm gives it the "unnamed alarm" fallback rather than letting it
	// through as a raw notification with a blank summary.
	var typeErr *json.UnmarshalTypeError
	isAlarm := (err == nil || errors.As(err, &typeErr)) && alarm.AlarmName != ""
	if isAlarm {
		return buildAlarm(alarm, region, topic)
	}

	return buildRaw(e, topic), cleanMeta(map[string]string{
		"topic":  topic,
		"region": region,
		"source": "sns-raw",
	}), true
}

func buildAlarm(al cloudWatchAlarm, region, topic string) (alert.Alert, map[string]string, bool) {
	if strings.EqualFold(al.NewStateValue, "INSUFFICIENT_DATA") {
		return alert.Alert{}, nil, false
	}

	// Test blank rather than empty: a whitespace-only AlarmName sanitizes to ""
	// and would otherwise produce a blank, unactionable summary.
	name := al.AlarmName
	if strings.TrimSpace(name) == "" {
		name = "unnamed alarm on " + topic
	}

	status := alert.StatusTriggered
	if al.NewStateValue == "OK" {
		status = alert.StatusClosed
	}

	return alert.Alert{
		Summary: validate.SanitizeText(name, alert.MaxSummaryLength),
		Details: validate.SanitizeText(alarmDetails(al, region, topic), alert.MaxDetailsLength),
		Source:  alert.SourceCloudwatch,
		Status:  status,

		// Cross-system contract: hex sha256 of the untruncated, unsanitized alarm
		// name, matching the CloudWatch alarm Lambdas that post to PagerDuty.
		// Changing it means one alarm produces two alerts.
		Dedup: alert.NewUserDedup(sha256Hex(name)),
	}, alarmMeta(al, region, topic), true
}

func alarmDetails(al cloudWatchAlarm, region, topic string) string {
	lines := []string{"State: " + al.OldStateValue + " -> " + al.NewStateValue}

	add := func(label, value string) {
		if value != "" {
			lines = append(lines, label+": "+value)
		}
	}
	add("Reason", truncRunes(al.NewStateReason, maxReasonLen))
	add("Changed", al.StateChangeTime)
	add("Region", region)
	add("Account", al.AWSAccountID)
	if al.Trigger.MetricName != "" {
		add("Metric", al.Trigger.Namespace+"/"+al.Trigger.MetricName)
	}
	add("Topic", topic)
	add("Alarm ARN", al.AlarmARN)

	if al.AlarmDescription != "" {
		// Blank line, then the description verbatim: it carries the runbook URL.
		lines = append(lines, "", al.AlarmDescription)
	}

	return strings.Join(lines, "\n")
}

func alarmMeta(al cloudWatchAlarm, region, topic string) map[string]string {
	return cleanMeta(map[string]string{
		"topic":       topic,
		"region":      region,
		"state":       al.NewStateValue,
		"aws_account": al.AWSAccountID,
		"namespace":   al.Trigger.Namespace,
		"metric":      al.Trigger.MetricName,
		"alarm_arn":   al.AlarmARN,
	})
}

// buildRaw maps a non-CloudWatch SNS notification. Raw notifications never close.
func buildRaw(e envelope, topic string) alert.Alert {
	// Each candidate is sanitized before the emptiness test: a whitespace-only
	// Subject is truthy but sanitizes to "".
	summary := sanitizeSummary(derefStr(e.Subject))
	if summary == "" {
		summary = sanitizeSummary(firstLine(e.Message))
	}
	if summary == "" {
		summary = sanitizeSummary("SNS notification on " + topic)
	}

	return alert.Alert{
		Summary: summary,
		Details: validate.SanitizeText(e.Message, alert.MaxDetailsLength),
		Source:  alert.SourceCloudwatch,
		Status:  alert.StatusTriggered,
		Dedup:   alert.NewUserDedup(sha256Hex(topic + "|" + summary)),
	}
}

// splitTopicARN returns the region and topic name from an SNS topic ARN of the
// form arn:aws:sns:<region>:<account>:<topic>. A malformed or short ARN yields
// empty segments rather than panicking.
func splitTopicARN(arn string) (region, topic string) {
	parts := strings.Split(arn, ":")
	if len(parts) > 3 {
		region = parts[3]
	}
	// strings.Split never returns an empty slice, so this index is always safe.
	return region, parts[len(parts)-1]
}

func sanitizeSummary(s string) string { return validate.SanitizeText(s, alert.MaxSummaryLength) }

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func truncRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// cleanMeta drops empty values and bounds the rest. Values are not sanitized:
// they are JSON-marshalled on write, so control characters are escaped rather
// than injected, and trimming would corrupt an ARN.
func cleanMeta(m map[string]string) map[string]string {
	for k, v := range m {
		if v == "" {
			delete(m, k)
			continue
		}
		m[k] = truncRunes(v, maxMetaValueLen)
	}
	return m
}
