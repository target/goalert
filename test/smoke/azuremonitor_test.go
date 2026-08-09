package smoke

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

const azAlertID = "/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/1db044ff-df8f-4064-a559-b9c9f5f4f000"

// azMetric builds a SingleResourceMultipleMetricCriteria delivery.
func azMetric(alertID, condition string) string {
	return fmt.Sprintf(`{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": %q,
      "alertRule": "Too Many Frontdoor Exceptions",
      "severity": "Sev2",
      "signalType": "Metric",
      "monitorCondition": %q,
      "monitoringService": "Platform",
      "alertTargetIDs": ["/subscriptions/sub-1/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/acct"],
      "configurationItems": ["stage-www-frontdoor"],
      "firedDateTime": "2026-07-18T08:01:01.000Z",
      "description": "Runbook: https://runbook.example.com/frontdoor"
    },
    "alertContext": {
      "conditionType": "SingleResourceMultipleMetricCriteria",
      "condition": {
        "windowSize": "PT5M",
        "allOf": [{
          "metricName": "Transactions",
          "operator": "GreaterThan",
          "threshold": "0",
          "timeAggregation": "Total",
          "dimensions": [{"name": "ApiName", "value": "GetBlob"}],
          "metricValue": 100
        }]
      }
    }
  }
}`, alertID, condition)
}

type azAlertNode struct {
	AlertID int
	Status  string
	Summary string
	Details string
	Meta    []struct {
		Key   string
		Value string
	}
}

// TestAzureMonitor exercises the Azure Monitor ingress end to end: alert
// creation, redelivery idempotency, the Resolved close and dedup-key release,
// the schema gate, token rejection, and body-size limits.
func TestAzureMonitor(t *testing.T) {
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
		({{uuid "int_key"}}, 'azureMonitor', 'my key', {{uuid "sid"}});
	`

	h := harness.NewHarness(t, sql, "azuremonitor-integration")
	defer h.Close()

	url := h.URL() + "/api/v2/azuremonitor/incoming?token=" + h.UUID("int_key")

	post := func(t *testing.T, body string) int {
		t.Helper()
		resp, err := http.Post(url, "application/json", strings.NewReader(body))
		require.NoError(t, err)
		resp.Body.Close()
		return resp.StatusCode
	}

	alerts := func(t *testing.T) []azAlertNode {
		t.Helper()
		res := h.GraphQLQuery2(`query{alerts(input:{includeNotified:true}){nodes{alertID status summary details meta{key value}}}}`)
		require.Empty(t, res.Errors)

		var result struct {
			Alerts struct{ Nodes []azAlertNode }
		}
		require.NoError(t, json.Unmarshal(res.Data, &result), "parse response: %s", string(res.Data))
		sort.Slice(result.Alerts.Nodes, func(i, j int) bool {
			return result.Alerts.Nodes[i].AlertID < result.Alerts.Nodes[j].AlertID
		})
		return result.Alerts.Nodes
	}

	// The UI is server-driven: the dropdown comes from integrationKeyTypes and the
	// copyable URL from IntegrationKey.href. If either misses the type, the
	// operator is handed a URL that does not work.
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
			if kt.ID == "azureMonitor" {
				found = true
				assert.Equal(t, "Azure Monitor", kt.Name)
				assert.True(t, kt.Enabled)
			}
		}
		assert.True(t, found, "azureMonitor must appear in integrationKeyTypes")

		res = h.GraphQLQuery2(`query{service(id:"` + h.UUID("sid") + `"){integrationKeys{type href}}}`)
		require.Empty(t, res.Errors)

		var svc struct {
			Service struct {
				IntegrationKeys []struct{ Type, Href string }
			}
		}
		require.NoError(t, json.Unmarshal(res.Data, &svc))
		require.Len(t, svc.Service.IntegrationKeys, 1)
		assert.Equal(t, "azureMonitor", svc.Service.IntegrationKeys[0].Type)
		assert.Contains(t, svc.Service.IntegrationKeys[0].Href, "/api/v2/azuremonitor/incoming")
	})

	var firstID int
	t.Run("metric alert creates an alert", func(t *testing.T) {
		require.Equal(t, http.StatusNoContent, post(t, azMetric(azAlertID, "Fired")))

		got := alerts(t)
		require.Len(t, got, 1)
		firstID = got[0].AlertID

		assert.Equal(t, "Too Many Frontdoor Exceptions", got[0].Summary)
		assert.Equal(t, "StatusUnacknowledged", got[0].Status)
		assert.Contains(t, got[0].Details, "Severity: Sev2")
		assert.Contains(t, got[0].Details, "Transactions GreaterThan 0 (Total, PT5M) = 100")
		assert.Contains(t, got[0].Details, "Runbook: https://runbook.example.com/frontdoor")
		// configurationItems preferred over the full ARM path.
		assert.Contains(t, got[0].Details, "Resource: stage-www-frontdoor")

		meta := map[string]string{}
		for _, m := range got[0].Meta {
			meta[m.Key] = m.Value
		}
		assert.Equal(t, "Metric", meta["signal_type"])
		assert.Equal(t, "Fired", meta["monitor_condition"])
		assert.Equal(t, azAlertID, meta["alert_id"])
	})

	t.Run("redelivery is idempotent", func(t *testing.T) {
		require.Equal(t, http.StatusNoContent, post(t, azMetric(azAlertID, "Fired")))

		got := alerts(t)
		require.Len(t, got, 1)
		assert.Equal(t, firstID, got[0].AlertID)
	})

	// autoMitigate is on for essentially every Azure metric rule, so this path
	// runs constantly in production.
	t.Run("resolved closes the alert", func(t *testing.T) {
		require.Equal(t, http.StatusNoContent, post(t, azMetric(azAlertID, "Resolved")))

		got := alerts(t)
		require.Len(t, got, 1)
		assert.Equal(t, firstID, got[0].AlertID)
		assert.Equal(t, "StatusClosed", got[0].Status)
	})

	// The close must free the dedup key, or every later firing of the same rule is
	// silently suppressed forever.
	t.Run("new firing after close creates a new alert", func(t *testing.T) {
		const nextFiring = "/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/3a10e1f4-0000-0000-0000-000000000000"
		require.Equal(t, http.StatusNoContent, post(t, azMetric(nextFiring, "Fired")))

		got := alerts(t)
		require.Len(t, got, 2)
		assert.NotEqual(t, firstID, got[1].AlertID)
		assert.Equal(t, "StatusUnacknowledged", got[1].Status)
	})

	// A Resolved delivery for an alert we never opened is normal, not an error.
	t.Run("resolved with no open alert", func(t *testing.T) {
		const unseen = "/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/never-seen"
		require.Equal(t, http.StatusNoContent, post(t, azMetric(unseen, "Resolved")))
		assert.Len(t, alerts(t), 2)
	})

	// Rejected with an actionable message rather than degraded to a blank alert.
	t.Run("legacy schema rejected with actionable message", func(t *testing.T) {
		body := `{"schemaId":"AzureMonitorMetricAlert","data":{"essentials":{"alertRule":"legacy"}}}`
		resp, err := http.Post(url, "application/json", strings.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		msg := make([]byte, 512)
		n, _ := resp.Body.Read(msg)
		assert.Contains(t, string(msg[:n]), "common alert schema")
		assert.Len(t, alerts(t), 2)
	})

	t.Run("service health payload is usable", func(t *testing.T) {
		body := `{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": "/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/sh-1",
      "alertRule": "ServiceHealthRule",
      "signalType": "Activity Log",
      "monitorCondition": "Fired",
      "monitoringService": "ServiceHealth"
    },
    "alertContext": {
      "properties": {
        "title": "Azure Storage degraded",
        "trackingId": "ABC-123",
        "communication": "<p>We are investigating</p>",
        "impactedServices": "[{\"ServiceName\":\"Storage\"}]"
      },
      "status": "Resolved"
    }
  }
}`
		require.Equal(t, http.StatusNoContent, post(t, body))

		got := alerts(t)
		require.Len(t, got, 3)
		assert.Equal(t, "ServiceHealthRule", got[2].Summary)
		assert.Contains(t, got[2].Details, "Tracking ID: ABC-123")
		assert.NotContains(t, got[2].Details, "<p>")
		// monitorCondition wins over alertContext.status.
		assert.Equal(t, "StatusUnacknowledged", got[2].Status)
	})

	t.Run("log alert carries its query and exactly one link", func(t *testing.T) {
		const link = "https://portal.azure.com/#blade/Microsoft_OperationalInsights/filtered-results"
		body := fmt.Sprintf(`{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": "/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/log-1",
      "alertRule": "Heartbeat missing",
      "severity": "Sev1",
      "signalType": "Log",
      "monitorCondition": "Fired",
      "monitoringService": "Log Alerts V2",
      "configurationItems": ["test-computer"]
    },
    "alertContext": {
      "conditionType": "LogQueryCriteria",
      "condition": {
        "windowSize": "PT10M",
        "allOf": [{
          "searchQuery": "Heartbeat | summarize count() by Computer",
          "metricMeasureColumn": null,
          "targetResourceTypes": "['Microsoft.OperationalInsights/workspaces']",
          "operator": "GreaterThan",
          "threshold": "0",
          "timeAggregation": "Count",
          "dimensions": [{"name": "Computer", "value": "test-computer"}],
          "metricValue": 3,
          "failingPeriods": {"numberOfEvaluationPeriods": 1, "minFailingPeriodsToAlert": 1},
          "linkToSearchResultsUI": "https://portal.azure.com/#unfiltered",
          "linkToFilteredSearchResultsUI": %q,
          "linkToSearchResultsAPI": "https://api.loganalytics.io/v1/unfiltered",
          "linkToFilteredSearchResultsAPI": "https://api.loganalytics.io/v1/filtered"
        }]
      }
    }
  }
}`, link)
		require.Equal(t, http.StatusNoContent, post(t, body))

		got := alerts(t)
		require.Len(t, got, 4)
		d := got[3].Details

		assert.Equal(t, "Heartbeat missing", got[3].Summary)
		assert.Contains(t, d, "Query: Heartbeat | summarize count() by Computer")
		assert.Contains(t, d, "GreaterThan 0 (Count, PT10M) = 3")
		assert.Contains(t, d, link)

		// Exactly one link, and neither *API variant: all four together exceed
		// the details limit on their own.
		assert.Equal(t, 1, strings.Count(d, "Results: "))
		assert.NotContains(t, d, "api.loganalytics.io")
		assert.NotContains(t, d, "#unfiltered")
	})

	t.Run("unrecognised conditionType still creates an alert", func(t *testing.T) {
		body := strings.Replace(
			azMetric("/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/webtest-1", "Fired"),
			`"conditionType": "SingleResourceMultipleMetricCriteria"`,
			`"conditionType": "WebtestLocationAvailabilityCriteria"`, 1)
		require.Equal(t, http.StatusNoContent, post(t, body))

		got := alerts(t)
		require.Len(t, got, 5)
		assert.NotEmpty(t, got[4].Summary)
	})

	t.Run("malformed body is a bad request", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, post(t, `{not json`))
		assert.Len(t, alerts(t), 5)
	})

	t.Run("oversized body is rejected", func(t *testing.T) {
		big := `{"schemaId":"azureMonitorCommonAlertSchema","data":{"essentials":{"alertRule":"` +
			strings.Repeat("x", 300*1024) + `"}}}`
		assert.Equal(t, http.StatusRequestEntityTooLarge, post(t, big))
		assert.Len(t, alerts(t), 5)
	})

	t.Run("wrong token rejected", func(t *testing.T) {
		bad := h.URL() + "/api/v2/azuremonitor/incoming?token=" + h.UUID("sid")
		resp, err := http.Post(bad, "application/json", strings.NewReader(azMetric(azAlertID, "Fired")))
		require.NoError(t, err)
		resp.Body.Close()

		assert.NotEqual(t, http.StatusNoContent, resp.StatusCode)
		assert.Len(t, alerts(t), 5)
	})

	t.Run("missing token rejected", func(t *testing.T) {
		resp, err := http.Post(h.URL()+"/api/v2/azuremonitor/incoming", "application/json",
			strings.NewReader(azMetric(azAlertID, "Fired")))
		require.NoError(t, err)
		resp.Body.Close()

		assert.NotEqual(t, http.StatusNoContent, resp.StatusCode)
		assert.Len(t, alerts(t), 5)
	})
}
