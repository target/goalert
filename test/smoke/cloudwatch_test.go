package smoke

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/app"
	"github.com/target/goalert/test/smoke/harness"
)

const cwTopicARN = "arn:aws:sns:us-west-2:123456789012:PagerDuty-Data"

// A genuinely allowlisted URL: only the transport is redirected to the test
// server, so the production host allowlist still runs against this string.
const cwCertURL = "https://sns.us-west-2.amazonaws.com/SimpleNotificationService-test.pem"

const cwAlarm = `{
	"AlarmName": "[us-west-2] Too Many Write Errors",
	"AlarmDescription": "Runbook: https://runbook.example.com/dp/ats-write-errors",
	"AWSAccountId": "123456789012",
	"NewStateValue": "%s",
	"OldStateValue": "OK",
	"NewStateReason": "Threshold Crossed",
	"StateChangeTime": "2026-07-30T00:00:00.000+0000",
	"AlarmArn": "arn:aws:cloudwatch:us-west-2:123456789012:alarm:x",
	"Region": "US West (Oregon)",
	"Trigger": {"Namespace": "Eightfold/DP", "MetricName": "WriteErrors"}
}`

// cwSignEnvelope signs the envelope the way AWS does: the signed fields in
// alphabetical order as name+"\n"+value+"\n", omitting Subject entirely when
// absent, then RSA PKCS#1 v1.5 over SHA-1 (SignatureVersion 1, AWS's default).
//
// This is a deliberate independent reimplementation of the production canonical
// string -- if the two ever disagree, that is exactly what this test should
// catch.
func cwSignEnvelope(t *testing.T, key *rsa.PrivateKey, env map[string]string) []byte {
	t.Helper()

	var fields []string
	switch env["Type"] {
	case "Notification":
		fields = []string{"Message", "MessageId", "Subject", "Timestamp", "TopicArn", "Type"}
	case "SubscriptionConfirmation", "UnsubscribeConfirmation":
		fields = []string{"Message", "MessageId", "SubscribeURL", "Timestamp", "Token", "TopicArn", "Type"}
	default:
		fields = []string{"Message", "MessageId", "Timestamp", "TopicArn", "Type"}
	}

	var sb strings.Builder
	for _, f := range fields {
		v, ok := env[f]
		if !ok {
			continue
		}
		sb.WriteString(f + "\n" + v + "\n")
	}

	sum := sha1.Sum([]byte(sb.String()))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, sum[:])
	require.NoError(t, err)

	env["SignatureVersion"] = "1"
	env["Signature"] = base64.StdEncoding.EncodeToString(sig)
	env["SigningCertURL"] = cwCertURL

	body, err := json.Marshal(env)
	require.NoError(t, err)

	return body
}

func cwNotification(message string) map[string]string {
	return map[string]string{
		"Type":      "Notification",
		"MessageId": "11111111-1111-1111-1111-111111111111",
		"TopicArn":  cwTopicARN,
		"Message":   message,
		"Timestamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	}
}

type cwAlertNode struct {
	AlertID int
	Status  string
	Summary string
	Details string
	Meta    []struct {
		Key   string
		Value string
	}
}

// TestCloudWatch exercises the full CloudWatch/SNS ingress path in-process: the
// subscription handshake, real RSA signature verification, the host allowlist,
// alarm mapping, dedup, and the OK close.
func TestCloudWatch(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');
	insert into integration_keys (id, type, name, service_id)
	values
		({{uuid "int_key"}}, 'cloudwatch', 'my key', {{uuid "sid"}});
	`

	signKey, certPEM := cwGenCert(t)
	forgedKey, _ := cwGenCert(t)

	certSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Action") == "ConfirmSubscription" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/x-pem-file")
		_, _ = w.Write(certPEM)
	}))
	defer certSrv.Close()

	h := harness.NewStoppedHarness(t, sql, nil, "cloudwatch-integration")
	defer h.Close()
	// The unsubscribe case logs at error level on purpose ("log loudly"), and the
	// harness fails a test on any error log line.
	h.IgnoreErrorsWith("subscription removed")
	h.StartWithAppCfgHook(func(c *app.Config) { c.CloudwatchBaseURL = certSrv.URL })

	url := h.URL() + "/api/v2/cloudwatch/incoming?token=" + h.UUID("int_key")

	// SNS always posts text/plain; the handler must parse regardless.
	post := func(t *testing.T, body []byte) int {
		t.Helper()
		resp, err := http.Post(url, "text/plain; charset=UTF-8", strings.NewReader(string(body)))
		require.NoError(t, err)
		resp.Body.Close()
		return resp.StatusCode
	}

	alerts := func(t *testing.T) []cwAlertNode {
		t.Helper()
		res := h.GraphQLQuery2(`query{alerts(input:{includeNotified:true}){nodes{alertID status summary details meta{key value}}}}`)
		require.Empty(t, res.Errors)

		var result struct {
			Alerts struct{ Nodes []cwAlertNode }
		}
		require.NoError(t, json.Unmarshal(res.Data, &result), "parse response: %s", string(res.Data))
		sort.Slice(result.Alerts.Nodes, func(i, j int) bool {
			return result.Alerts.Nodes[i].AlertID < result.Alerts.Nodes[j].AlertID
		})
		return result.Alerts.Nodes
	}

	// The UI is entirely server-driven: the dropdown comes from
	// integrationKeyTypes and the copy-to-clipboard value comes from
	// IntegrationKey.href. If either is missing the key type, the user is handed a
	// URL that silently will not work with SNS.
	t.Run("ui offers the type and the right url", func(t *testing.T) {
		res := h.GraphQLQuery2(`query{integrationKeyTypes{id name label enabled}}`)
		require.Empty(t, res.Errors)

		var types struct {
			IntegrationKeyTypes []struct {
				ID, Name, Label string
				Enabled         bool
			}
		}
		require.NoError(t, json.Unmarshal(res.Data, &types))

		var found bool
		for _, kt := range types.IntegrationKeyTypes {
			if kt.ID != "cloudwatch" {
				continue
			}
			found = true
			assert.Equal(t, "Amazon CloudWatch", kt.Name)
			assert.Equal(t, "CloudWatch Webhook URL", kt.Label)
			assert.True(t, kt.Enabled)
		}
		assert.True(t, found, "cloudwatch must appear in integrationKeyTypes")

		res = h.GraphQLQuery2(`query{service(id:"` + h.UUID("sid") + `"){integrationKeys{id type href}}}`)
		require.Empty(t, res.Errors)

		var svc struct {
			Service struct {
				IntegrationKeys []struct{ ID, Type, Href string }
			}
		}
		require.NoError(t, json.Unmarshal(res.Data, &svc))
		require.Len(t, svc.Service.IntegrationKeys, 1)

		key := svc.Service.IntegrationKeys[0]
		assert.Equal(t, "cloudwatch", key.Type)
		assert.Contains(t, key.Href, "/api/v2/cloudwatch/incoming")
		assert.Contains(t, key.Href, "token="+h.UUID("int_key"))
	})

	t.Run("subscription confirmation", func(t *testing.T) {
		env := map[string]string{
			"Type":         "SubscriptionConfirmation",
			"MessageId":    "22222222-2222-2222-2222-222222222222",
			"TopicArn":     cwTopicARN,
			"Message":      "You have chosen to subscribe to the topic",
			"Timestamp":    time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			"Token":        "confirm-token",
			"SubscribeURL": "https://sns.us-west-2.amazonaws.com/?Action=ConfirmSubscription&TopicArn=" + cwTopicARN + "&Token=confirm-token",
		}
		assert.Equal(t, http.StatusOK, post(t, cwSignEnvelope(t, signKey, env)))
		assert.Empty(t, alerts(t))
	})

	var alarmID int
	t.Run("alarm creates an alert", func(t *testing.T) {
		body := cwSignEnvelope(t, signKey, cwNotification(fmt.Sprintf(cwAlarm, "ALARM")))
		require.Equal(t, http.StatusNoContent, post(t, body))

		got := alerts(t)
		require.Len(t, got, 1)
		alarmID = got[0].AlertID

		assert.Equal(t, "[us-west-2] Too Many Write Errors", got[0].Summary)
		assert.Contains(t, got[0].Details, "State: OK -> ALARM")
		assert.Contains(t, got[0].Details, "Runbook: https://runbook.example.com/dp/ats-write-errors")
		// Region must come from the topic ARN, not CloudWatch's display name.
		assert.Contains(t, got[0].Details, "Region: us-west-2")
		assert.NotContains(t, got[0].Details, "Oregon")

		meta := map[string]string{}
		for _, m := range got[0].Meta {
			meta[m.Key] = m.Value
		}
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

	t.Run("redelivery is idempotent", func(t *testing.T) {
		body := cwSignEnvelope(t, signKey, cwNotification(fmt.Sprintf(cwAlarm, "ALARM")))
		require.Equal(t, http.StatusNoContent, post(t, body))

		got := alerts(t)
		require.Len(t, got, 1)
		assert.Equal(t, alarmID, got[0].AlertID)
	})

	t.Run("insufficient data creates nothing", func(t *testing.T) {
		body := cwSignEnvelope(t, signKey, cwNotification(fmt.Sprintf(cwAlarm, "INSUFFICIENT_DATA")))
		require.Equal(t, http.StatusNoContent, post(t, body))
		assert.Len(t, alerts(t), 1)
	})

	t.Run("ok closes the alert", func(t *testing.T) {
		body := cwSignEnvelope(t, signKey, cwNotification(fmt.Sprintf(cwAlarm, "OK")))
		require.Equal(t, http.StatusNoContent, post(t, body))

		got := alerts(t)
		require.Len(t, got, 1)
		assert.Equal(t, alarmID, got[0].AlertID)
		assert.Equal(t, "StatusClosed", got[0].Status)

		// Metadata is written only when the alert is new, so state stays at the
		// value it had at creation. Asserting otherwise would be wrong.
		meta := map[string]string{}
		for _, m := range got[0].Meta {
			meta[m.Key] = m.Value
		}
		assert.Equal(t, "ALARM", meta["state"])
	})

	// The created alert is nil with a nil error here; a 500 or an error log would
	// fail this test twice over.
	t.Run("stray ok with no open alert", func(t *testing.T) {
		body := cwSignEnvelope(t, signKey, cwNotification(`{"AlarmName":"never-seen","NewStateValue":"OK"}`))
		require.Equal(t, http.StatusNoContent, post(t, body))
		assert.Len(t, alerts(t), 1)
	})

	t.Run("raw notification uses subject", func(t *testing.T) {
		env := cwNotification("this is not json")
		env["Subject"] = "Backup failed"
		require.Equal(t, http.StatusNoContent, post(t, cwSignEnvelope(t, signKey, env)))

		got := alerts(t)
		require.Len(t, got, 2)
		assert.Equal(t, "Backup failed", got[1].Summary)

		meta := map[string]string{}
		for _, m := range got[1].Meta {
			meta[m.Key] = m.Value
		}
		assert.Equal(t, "sns-raw", meta["source"])
	})

	t.Run("forged signature is rejected", func(t *testing.T) {
		body := cwSignEnvelope(t, forgedKey, cwNotification(`{"AlarmName":"forged","NewStateValue":"ALARM"}`))
		assert.Equal(t, http.StatusForbidden, post(t, body))
		assert.Len(t, alerts(t), 2)
	})

	// Proves the allowlist is still live even with CloudwatchBaseURL set.
	t.Run("cert host not allowlisted", func(t *testing.T) {
		env := cwNotification(`{"AlarmName":"evil","NewStateValue":"ALARM"}`)
		body := cwSignEnvelope(t, signKey, env)

		var raw map[string]string
		require.NoError(t, json.Unmarshal(body, &raw))
		raw["SigningCertURL"] = "https://sns.us-west-2.amazonaws.com.evil.com/x.pem"
		tampered, err := json.Marshal(raw)
		require.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, post(t, tampered))
		assert.Len(t, alerts(t), 2)
	})

	t.Run("unknown type is a bad request", func(t *testing.T) {
		env := cwNotification(`{"AlarmName":"x","NewStateValue":"ALARM"}`)
		env["Type"] = "Bogus"
		assert.Equal(t, http.StatusBadRequest, post(t, cwSignEnvelope(t, signKey, env)))
		assert.Len(t, alerts(t), 2)
	})

	t.Run("unsubscribe confirmation", func(t *testing.T) {
		env := map[string]string{
			"Type":         "UnsubscribeConfirmation",
			"MessageId":    "33333333-3333-3333-3333-333333333333",
			"TopicArn":     cwTopicARN,
			"Message":      "You have chosen to deactivate subscription",
			"Timestamp":    time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			"Token":        "unsub-token",
			"SubscribeURL": "https://sns.us-west-2.amazonaws.com/?Action=ConfirmSubscription&Token=unsub-token",
		}
		assert.Equal(t, http.StatusOK, post(t, cwSignEnvelope(t, signKey, env)))
		assert.Len(t, alerts(t), 2)
	})
}

// cwGenCert returns an RSA key and a self-signed PEM certificate for it. Only the
// public key is ever read back, so no chain is needed.
func cwGenCert(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "sns.us-west-2.amazonaws.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	return key, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}
