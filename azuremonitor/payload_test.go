package azuremonitor

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/alert"
)

const (
	testAlertID   = "/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/1db044ff-df8f-4064-a559-b9c9f5f4f000"
	testServiceID = "3c1a1a44-8e7a-4d1b-9f4a-2b0e5c6d7f80"
	testRunbook   = "https://runbook.example.com/frontdoor"
)

// metricPayload is a SingleResourceMultipleMetricCriteria delivery.
func metricPayload(condition string) string {
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
      "originAlertId": "sub-1_rg_Microsoft.Insights_metricAlerts_rule_604016583",
      "firedDateTime": "2026-07-18T08:01:01.000Z",
      "description": "Runbook: %s"
    },
    "alertContext": {
      "conditionType": "SingleResourceMultipleMetricCriteria",
      "condition": {
        "windowSize": "PT5M",
        "allOf": [{
          "metricName": "Transactions",
          "metricNamespace": "Microsoft.Storage/storageAccounts",
          "operator": "GreaterThan",
          "threshold": "0",
          "timeAggregation": "Total",
          "dimensions": [{"name": "ApiName", "value": "GetBlob"}],
          "metricValue": 100,
          "webTestName": null
        }],
        "windowStartTime": "2026-07-18T07:56:01.000Z",
        "windowEndTime": "2026-07-18T08:01:01.000Z"
      }
    },
    "customProperties": {"runbook": %q}
  }
}`, testAlertID, condition, testRunbook, testRunbook)
}

func TestBuildAlert_Metric(t *testing.T) {
	t.Run("golden fired alert", func(t *testing.T) {
		a, meta, _, err := buildAlert([]byte(metricPayload("Fired")))
		require.NoError(t, err)

		assert.Equal(t, "Too Many Frontdoor Exceptions", a.Summary)
		assert.Equal(t, alert.StatusTriggered, a.Status)
		assert.Equal(t, alert.SourceAzureMonitor, a.Source)

		require.NotNil(t, a.Dedup)
		assert.Equal(t, sha256Hex(testAlertID), a.Dedup.Payload)

		assert.Contains(t, a.Details, "Severity: Sev2")
		assert.Contains(t, a.Details, "Signal type: Metric")
		// configurationItems is preferred over the full ARM path.
		assert.Contains(t, a.Details, "Resource: stage-www-frontdoor")
		assert.NotContains(t, a.Details, "/resourceGroups/rg/providers")
		assert.Contains(t, a.Details, "Transactions GreaterThan 0 (Total, PT5M) = 100")
		assert.Contains(t, a.Details, "Dimensions: ApiName=GetBlob")
		assert.Contains(t, a.Details, testRunbook)

		assert.Equal(t, "Metric", meta["signal_type"])
		assert.Equal(t, "Fired", meta["monitor_condition"])
		assert.Equal(t, testAlertID, meta["alert_id"])
		assert.Equal(t, "stage-www-frontdoor", meta["configuration_items"])
	})

	// The close path is live for Azure (autoMitigate), so the Resolved delivery
	// must produce the same dedup key or the alert never closes.
	t.Run("resolved closes with the same dedup", func(t *testing.T) {
		fired, _, _, err := buildAlert([]byte(metricPayload("Fired")))
		require.NoError(t, err)
		resolved, meta, _, err := buildAlert([]byte(metricPayload("Resolved")))
		require.NoError(t, err)

		assert.Equal(t, alert.StatusClosed, resolved.Status)
		require.NotNil(t, resolved.Dedup)
		assert.Equal(t, fired.Dedup.Payload, resolved.Dedup.Payload)
		assert.Equal(t, "Resolved", meta["monitor_condition"])
	})

	// originAlertId is per-rule for metric alerts; using it would mute the rule
	// forever after one missed close.
	t.Run("dedup ignores originAlertId", func(t *testing.T) {
		a, _, _, err := buildAlert([]byte(metricPayload("Fired")))
		require.NoError(t, err)
		require.NotNil(t, a.Dedup)
		assert.NotEqual(t, sha256Hex("sub-1_rg_Microsoft.Insights_metricAlerts_rule_604016583"), a.Dedup.Payload)
	})

	t.Run("blank alertRule falls back to resource", func(t *testing.T) {
		body := strings.Replace(metricPayload("Fired"), `"alertRule": "Too Many Frontdoor Exceptions"`, `"alertRule": "   "`, 1)
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)
		assert.Equal(t, "Azure Monitor alert on stage-www-frontdoor", a.Summary)
	})

	t.Run("description preserved verbatim", func(t *testing.T) {
		a, _, _, err := buildAlert([]byte(metricPayload("Fired")))
		require.NoError(t, err)
		assert.Contains(t, a.Details, "Runbook: "+testRunbook)
	})

	t.Run("falls back to alertTargetIDs when no configurationItems", func(t *testing.T) {
		body := strings.Replace(metricPayload("Fired"), `"configurationItems": ["stage-www-frontdoor"],`, "", 1)
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)
		assert.Contains(t, a.Details, "Resource: /subscriptions/sub-1/resourceGroups/rg")
	})

	// allOf is an array; a rule can carry several criteria and all must render.
	t.Run("renders every allOf entry", func(t *testing.T) {
		body := strings.Replace(metricPayload("Fired"),
			`"metricValue": 100,
          "webTestName": null
        }]`,
			`"metricValue": 100,
          "webTestName": null
        }, {
          "metricName": "Latency",
          "operator": "GreaterThan",
          "threshold": "500",
          "timeAggregation": "Average",
          "dimensions": [],
          "metricValue": 900
        }]`, 1)
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)
		assert.Contains(t, a.Details, "Transactions GreaterThan 0")
		assert.Contains(t, a.Details, "Latency GreaterThan 500 (Average, PT5M) = 900")
	})

	// threshold is a string and metricValue a number -- neither may be assumed to
	// be the other's type.
	t.Run("integral metricValue renders without a decimal", func(t *testing.T) {
		a, _, _, err := buildAlert([]byte(metricPayload("Fired")))
		require.NoError(t, err)
		assert.Contains(t, a.Details, "= 100")
		assert.NotContains(t, a.Details, "= 100.0")
	})
}

func dynamicPayload() string {
	return fmt.Sprintf(`{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": %q,
      "alertRule": "Dynamic Transactions",
      "severity": "Sev3",
      "signalType": "Metric",
      "monitorCondition": "Fired",
      "monitoringService": "Platform",
      "configurationItems": ["test-storageAccount"]
    },
    "alertContext": {
      "conditionType": "DynamicThresholdCriteria",
      "condition": {
        "windowSize": "PT15M",
        "allOf": [{
          "alertSensitivity": "Low",
          "failingPeriods": {"numberOfEvaluationPeriods": 3, "minFailingPeriodsToAlert": 3},
          "ignoreDataBefore": null,
          "metricName": "Transactions",
          "operator": "GreaterThan",
          "threshold": "0.3",
          "timeAggregation": "Average",
          "dimensions": [],
          "metricValue": 78.09
        }]
      }
    }
  }
}`, testAlertID)
}

func TestBuildAlert_DynamicThreshold(t *testing.T) {
	a, _, _, err := buildAlert([]byte(dynamicPayload()))
	require.NoError(t, err)

	// The threshold is a sensitivity artifact, not a limit -- rendering "0.3"
	// would mislead whoever is paged.
	assert.NotContains(t, a.Details, "0.3")
	assert.Contains(t, a.Details, "Sensitivity: Low")
	assert.Contains(t, a.Details, "Failing periods: 3 of 3")
	assert.Contains(t, a.Details, "Transactions GreaterThan (Average, PT15M) = 78.09")
}

func logPayload(searchQuery, link string) string {
	return fmt.Sprintf(`{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": %q,
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
          "searchQuery": %q,
          "metricMeasureColumn": null,
          "targetResourceTypes": "['Microsoft.OperationalInsights/workspaces']",
          "operator": "GreaterThan",
          "threshold": "0",
          "timeAggregation": "Count",
          "dimensions": [{"name": "Computer", "value": "test-computer"}],
          "metricValue": 3,
          "failingPeriods": {"numberOfEvaluationPeriods": 1, "minFailingPeriodsToAlert": 1},
          "linkToSearchResultsUI": "https://portal.azure.com#@unfiltered",
          "linkToFilteredSearchResultsUI": %q,
          "linkToSearchResultsAPI": "https://api.loganalytics.io/v1/workspaces/unfiltered",
          "linkToFilteredSearchResultsAPI": "https://api.loganalytics.io/v1/workspaces/filtered"
        }]
      }
    }
  }
}`, testAlertID, searchQuery, link)
}

func TestBuildAlert_LogQuery(t *testing.T) {
	const link = "https://portal.azure.com#@filtered-link"

	t.Run("golden log alert", func(t *testing.T) {
		a, meta, _, err := buildAlert([]byte(logPayload("Heartbeat", link)))
		require.NoError(t, err)

		assert.Equal(t, "Heartbeat missing", a.Summary)
		assert.Equal(t, alert.StatusTriggered, a.Status)
		assert.Contains(t, a.Details, "Query: Heartbeat")
		assert.Contains(t, a.Details, "GreaterThan 0 (Count, PT10M) = 3")
		assert.Contains(t, a.Details, "Failing periods: 1 of 1")
		assert.Contains(t, a.Details, "Dimensions: Computer=test-computer")
		assert.Equal(t, "Log", meta["signal_type"])
	})

	// Exactly one link, and the filtered UI one. All four together exceed
	// MaxDetailsLength on their own.
	t.Run("exactly one link, the filtered UI one", func(t *testing.T) {
		a, _, _, err := buildAlert([]byte(logPayload("Heartbeat", link)))
		require.NoError(t, err)

		assert.Contains(t, a.Details, link)
		assert.NotContains(t, a.Details, "unfiltered")
		assert.NotContains(t, a.Details, "api.loganalytics.io")
		assert.Equal(t, 1, strings.Count(a.Details, "Results: "))
	})

	t.Run("falls back to unfiltered link when filtered is absent", func(t *testing.T) {
		body := strings.Replace(logPayload("Heartbeat", link), `"linkToFilteredSearchResultsUI": "`+link+`",`, "", 1)
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)
		assert.Contains(t, a.Details, "https://portal.azure.com#@unfiltered")
	})

	t.Run("metricMeasureColumn rendered only when set", func(t *testing.T) {
		a, _, _, err := buildAlert([]byte(logPayload("Heartbeat", link)))
		require.NoError(t, err)
		assert.NotContains(t, a.Details, "Measure column")

		body := strings.Replace(logPayload("Heartbeat", link), `"metricMeasureColumn": null`, `"metricMeasureColumn": "Duration"`, 1)
		a, _, _, err = buildAlert([]byte(body))
		require.NoError(t, err)
		assert.Contains(t, a.Details, "Measure column: Duration")
	})

	// targetResourceTypes is a string containing a JSON array, not an array.
	t.Run("string-containing-json targetResourceTypes does not error", func(t *testing.T) {
		a, _, _, err := buildAlert([]byte(logPayload("Heartbeat", link)))
		require.NoError(t, err)
		assert.NotEmpty(t, a.Summary)
	})

	// The test that catches naive truncation: the link must survive whole, not be
	// cut mid-URL, even when the query preceding it is enormous.
	t.Run("long query keeps the link intact", func(t *testing.T) {
		longQuery := strings.Repeat("Heartbeat | where Computer == 'x' | summarize count() ", 200)
		longLink := "https://portal.azure.com#@" + strings.Repeat("q", 1000)

		a, _, _, err := buildAlert([]byte(logPayload(longQuery, longLink)))
		require.NoError(t, err)

		assert.LessOrEqual(t, len([]rune(a.Details)), alert.MaxDetailsLength)
		assert.Contains(t, a.Details, longLink, "the full link must survive truncation")
		// The query is capped so it cannot crowd the link out.
		assert.NotContains(t, a.Details, strings.Repeat("Heartbeat | where Computer == 'x' | summarize count() ", 20))
	})
}

func serviceHealthPayload() string {
	return fmt.Sprintf(`{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": %q,
      "alertRule": "test-ServiceHealthAlertRule",
      "severity": "Sev4",
      "signalType": "Activity Log",
      "monitorCondition": "Fired",
      "monitoringService": "ServiceHealth",
      "alertTargetIDs": ["/subscriptions/sub-1"]
    },
    "alertContext": {
      "authorization": null, "channels": 1, "claims": null, "caller": null,
      "eventSource": 2, "level": 3,
      "operationName": "Microsoft.ServiceHealth/incident/action",
      "properties": {
        "title": "Test Action Group - Test Service Health Alert",
        "service": "Azure Service Name",
        "region": "Global",
        "communication": "<p>This is a test from Service Health Alert</p>",
        "incidentType": "Incident",
        "trackingId": "TEST-TTT",
        "impactStartTime": "2026-07-31T13:00:00.000Z",
        "impactedServices": "[{\"ImpactedRegions\":[{\"RegionName\":\"Global\"}],\"ServiceName\":\"Azure Service Name\"}]",
        "stage": "Resolved",
        "isHIR": "false"
      },
      "status": "Resolved", "subStatus": null, "ResourceType": null
    }
  }
}`, testAlertID)
}

// Service Health carries no conditionType at all, which is why conditionType
// dispatch must tolerate its absence.
func TestBuildAlert_ServiceHealth(t *testing.T) {
	a, meta, _, err := buildAlert([]byte(serviceHealthPayload()))
	require.NoError(t, err)

	assert.Equal(t, "test-ServiceHealthAlertRule", a.Summary)
	assert.Contains(t, a.Details, "Title: Test Action Group - Test Service Health Alert")
	assert.Contains(t, a.Details, "Tracking ID: TEST-TTT")
	assert.Contains(t, a.Details, "Service: Azure Service Name")
	assert.Contains(t, a.Details, "Stage: Resolved")

	// HTML is unreadable on a pager; tags are stripped.
	assert.NotContains(t, a.Details, "<p>")
	assert.NotContains(t, a.Details, "</p>")

	// monitorCondition wins over alertContext.status: the incident resolved while
	// the alert fired. Service Health is where this trap actually bites.
	assert.Equal(t, alert.StatusTriggered, a.Status)
	assert.Equal(t, "Fired", meta["monitor_condition"])
}

func TestBuildAlert_SchemaGate(t *testing.T) {
	tests := []struct{ name, schemaID string }{
		{name: "missing", schemaID: ""},
		{name: "legacy metric", schemaID: "AzureMonitorMetricAlert"},
		{name: "unrecognised", schemaID: "somethingElse"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"schemaId": %q, "data": {"essentials": {"alertRule": "x"}}}`, tt.schemaID)
			_, _, _, err := buildAlert([]byte(body))

			// Rejected with an actionable message, not silently degraded to the
			// best-effort path.
			require.ErrorIs(t, err, errLegacySchema)
			assert.Contains(t, err.Error(), "common alert schema")
		})
	}
}

// Multi-resource and resource-group-scoped metric rules emit
// MultipleResourceMultipleMetricCriteria, not SingleResource... . Ten of this
// tenant's metric rules use it, and before the envelope-based dispatch they
// rendered essentials only -- silently losing every metric field.
func TestBuildAlert_MultipleResourceMetric(t *testing.T) {
	body := strings.Replace(metricPayload("Fired"),
		`"conditionType": "SingleResourceMultipleMetricCriteria"`,
		`"conditionType": "MultipleResourceMultipleMetricCriteria"`, 1)

	a, _, _, err := buildAlert([]byte(body))
	require.NoError(t, err)

	assert.Contains(t, a.Details, "Transactions GreaterThan 0 (Total, PT5M) = 100")
	assert.Contains(t, a.Details, "Namespace: Microsoft.Storage/storageAccounts")
	assert.Contains(t, a.Details, "Dimensions: ApiName=GetBlob")
}

func webtestPayload() string {
	return `{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": "wt-1", "alertRule": "availability-check",
      "monitorCondition": "Fired", "signalType": "Metric", "monitoringService": "Platform"
    },
    "alertContext": {
      "conditionType": "WebtestLocationAvailabilityCriteria",
      "condition": {"windowSize": "PT5M", "allOf": [{
        "metricName": "Failed Location", "metricNamespace": null,
        "operator": "GreaterThan", "threshold": "2", "timeAggregation": "Sum",
        "dimensions": [], "metricValue": 5,
        "webTestName": "myAvailabilityTest-myApplication"
      }]}
    }
  }
}`
}

func TestBuildAlert_WebtestAvailability(t *testing.T) {
	a, _, _, err := buildAlert([]byte(webtestPayload()))
	require.NoError(t, err)

	assert.Contains(t, a.Details, "Failed Location GreaterThan 2 (Sum, PT5M) = 5")
	assert.Contains(t, a.Details, "Web test: myAvailabilityTest-myApplication")
}

// Azure documents threshold and dimension values as strings but is not
// consistent across shapes. A hard unmarshal failure would 400, which Azure does
// not retry, so one oddly-typed field must not cost the whole page.
func TestBuildAlert_ToleratesTypeMismatch(t *testing.T) {
	t.Run("numeric threshold", func(t *testing.T) {
		body := strings.Replace(metricPayload("Fired"), `"threshold": "0"`, `"threshold": 25`, 1)
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)

		// The mistyped value is lost, everything else survives.
		assert.Contains(t, a.Details, "Transactions GreaterThan")
		assert.Contains(t, a.Details, "Resource: stage-www-frontdoor")
		require.NotNil(t, a.Dedup)
		assert.Equal(t, sha256Hex(testAlertID), a.Dedup.Payload)
	})

	t.Run("numeric dimension value", func(t *testing.T) {
		body := strings.Replace(metricPayload("Fired"),
			`"dimensions": [{"name": "ApiName", "value": "GetBlob"}]`,
			`"dimensions": [{"name": "code", "value": 500}]`, 1)
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)

		assert.NotEmpty(t, a.Summary)
		require.NotNil(t, a.Dedup)
	})
}

// essentials fields that apply to every signal type, so they also improve the
// fallback path where they are the only content available.
func TestBuildAlert_EssentialsExtras(t *testing.T) {
	body := strings.Replace(metricPayload("Fired"),
		`"description": "Runbook: `+testRunbook+`"`,
		`"description": "Runbook: `+testRunbook+`",
		 "alertRuleId": "/subscriptions/sub-1/resourceGroups/rg/providers/microsoft.insights/metricAlerts/rule",
		 "targetResourceGroup": "stage-rg",
		 "targetResourceType": "Microsoft.Storage/storageAccounts",
		 "investigationLink": "https://portal.azure.com/investigate/abc"`, 1)

	a, meta, _, err := buildAlert([]byte(body))
	require.NoError(t, err)

	assert.Contains(t, a.Details, "Resource group: stage-rg")
	assert.Contains(t, a.Details, "Resource type: Microsoft.Storage/storageAccounts")
	assert.Contains(t, a.Details, "Investigate: https://portal.azure.com/investigate/abc")

	// The rule's ARM ID is machine-facing, so it belongs in metadata rather than
	// cluttering the pager text with a full resource path.
	assert.Equal(t, "stage-rg", meta["target_resource_group"])
	assert.Equal(t, "Microsoft.Storage/storageAccounts", meta["target_resource_type"])
	assert.Contains(t, meta["alert_rule_id"], "metricAlerts/rule")
}

// Azure Managed Prometheus rule groups route to PagerDuty in this tenant (three
// "Azure Pod Health Degraded" groups). The shape has no conditionType and no
// condition.allOf, so before the dedicated branch it rendered essentials only.
func prometheusPayload() string {
	return `{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": "/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/prom-1",
      "alertRule": "Azure Pod Health Degraded (microservices-prod)",
      "severity": "Sev2", "signalType": "Metric",
      "monitorCondition": "Fired", "monitoringService": "Prometheus",
      "configurationItems": ["aks-prod"]
    },
    "alertContext": {
      "interval": "PT1M",
      "expression": "kube_pod_status_ready{condition=\"false\"} > 0",
      "expressionValue": "3",
      "for": "PT5M",
      "labels": {"cluster": "microservices-prod", "namespace": "default", "severity": "warning"},
      "annotations": {
        "summary": "Pods are not ready in microservices-prod",
        "description": "3 pods have been NotReady for more than 5 minutes",
        "runbook_url": "https://runbook.example.com/pods"
      },
      "ruleGroup": "/subscriptions/sub-1/resourceGroups/rg/providers/Microsoft.AlertsManagement/prometheusRuleGroups/pod-health"
    }
  }
}`
}

func TestBuildAlert_Prometheus(t *testing.T) {
	a, meta, _, err := buildAlert([]byte(prometheusPayload()))
	require.NoError(t, err)

	assert.Equal(t, "Azure Pod Health Degraded (microservices-prod)", a.Summary)
	assert.Equal(t, alert.StatusTriggered, a.Status)
	require.NotNil(t, a.Dedup)

	// Annotations lead, as prose, because they are author-written.
	assert.Contains(t, a.Details, "Pods are not ready in microservices-prod")
	assert.Contains(t, a.Details, "3 pods have been NotReady for more than 5 minutes")
	assert.Contains(t, a.Details, "runbook_url: https://runbook.example.com/pods")

	assert.Contains(t, a.Details, `Expression: kube_pod_status_ready{condition="false"} > 0`)
	assert.Contains(t, a.Details, "Value: 3")
	assert.Contains(t, a.Details, "For: PT5M")
	assert.Contains(t, a.Details, "Interval: PT1M")
	// Labels are Prometheus's dimensions, rendered in a stable order.
	assert.Contains(t, a.Details, "Labels: cluster=microservices-prod, namespace=default, severity=warning")

	assert.Equal(t, "Prometheus", meta["monitoring_service"])
	assert.Contains(t, meta["rule_group"], "prometheusRuleGroups/pod-health")

	// Must not be mistaken for a metric or log alert.
	assert.NotContains(t, a.Details, "Query:")
	assert.NotContains(t, a.Details, "Namespace:")
}

// Microsoft's documented Prometheus sample, verbatim, with no annotations set.
func TestBuildAlert_PrometheusDocSample(t *testing.T) {
	body := `{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {"alertId": "p1", "alertRule": "sql-availability",
      "monitorCondition": "Fired", "signalType": "Metric", "monitoringService": "Prometheus"},
    "alertContext": {
      "interval": "PT1M", "expression": "sql_up > 0", "expressionValue": "0", "for": "PT2M",
      "labels": {"Environment": "Prod", "cluster": "myCluster1"},
      "annotations": {"summary": "alert on SQL availability"},
      "ruleGroup": "/subscriptions/s/resourceGroups/rg/providers/Microsoft.AlertsManagement/prometheusRuleGroups/g"
    }
  }
}`
	a, _, _, err := buildAlert([]byte(body))
	require.NoError(t, err)

	assert.Contains(t, a.Details, "alert on SQL availability")
	assert.Contains(t, a.Details, "Expression: sql_up > 0")
	assert.Contains(t, a.Details, "Value: 0")
	assert.Contains(t, a.Details, "Labels: Environment=Prod, cluster=myCluster1")
}

// parseInfo is what makes an unrecognised payload visible in the logs. Without
// ContextRendered the only symptom of a newly-routed Azure alert type is a thin
// alert nobody notices.
func TestBuildAlert_ParseInfo(t *testing.T) {
	tests := []struct {
		name              string
		body              string
		wantService       string
		wantConditionType string
		wantRendered      bool
	}{
		{
			name: "static metric", body: metricPayload("Fired"),
			wantService: "Platform", wantConditionType: "SingleResourceMultipleMetricCriteria",
			wantRendered: true,
		},
		{
			name: "log alerts v2", body: logPayload("Heartbeat", "https://portal.azure.com#@x"),
			wantConditionType: "LogQueryCriteria", wantRendered: true,
		},
		{
			// No conditionType at all, but still recognised via its own fields.
			name: "prometheus", body: prometheusPayload(),
			wantService: "Prometheus", wantConditionType: "", wantRendered: true,
		},
		{
			name: "service health", body: serviceHealthPayload(),
			wantService: "ServiceHealth", wantConditionType: "", wantRendered: true,
		},
		{
			// The case the log line exists for.
			name: "unrecognised shape",
			body: `{"schemaId":"azureMonitorCommonAlertSchema","data":{
				"essentials":{"alertId":"b1","alertRule":"BackupJobFailed",
				  "signalType":"Log","monitoringService":"Azure Backup","monitorCondition":"Fired"},
				"alertContext":{"BackupItemName":"vm-1","JobFailureCode":"UserErrorX"}}}`,
			wantService: "Azure Backup", wantConditionType: "", wantRendered: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, info, err := buildAlert([]byte(tt.body))
			require.NoError(t, err)

			assert.Equal(t, tt.wantRendered, info.ContextRendered, "ContextRendered")
			assert.Equal(t, tt.wantConditionType, info.ConditionType, "ConditionType")
			if tt.wantService != "" {
				assert.Equal(t, tt.wantService, info.MonitorService, "MonitorService")
			}
		})
	}
}

func TestBuildAlert_Fallback(t *testing.T) {
	// An unknown conditionType that still carries condition.allOf is rendered, not
	// dropped -- dispatch is on the envelope, not on a name allowlist. Only a
	// payload with no condition at all falls back to essentials.
	t.Run("unknown conditionType with a condition still renders", func(t *testing.T) {
		body := strings.Replace(metricPayload("Fired"),
			`"conditionType": "SingleResourceMultipleMetricCriteria"`,
			`"conditionType": "SomeFutureAzureCriteria"`, 1)
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)

		assert.Equal(t, "Too Many Frontdoor Exceptions", a.Summary)
		assert.Contains(t, a.Details, "Transactions GreaterThan 0 (Total, PT5M) = 100")
		require.NotNil(t, a.Dedup)
	})

	// signalType Log with a non-log monitoringService must not go through the KQL
	// parser -- this is why conditionType, not signalType, is the discriminator.
	t.Run("azure backup under signalType Log", func(t *testing.T) {
		body := `{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {
      "alertId": "/subscriptions/sub-1/providers/Microsoft.AlertsManagement/alerts/backup-1",
      "alertRule": "BackupJobFailed",
      "signalType": "Log",
      "monitorCondition": "Fired",
      "monitoringService": "Azure Backup"
    },
    "alertContext": {"BackupItemName": "vm-1", "JobFailureCode": "UserErrorX"}
  }
}`
		a, meta, _, err := buildAlert([]byte(body))
		require.NoError(t, err)

		assert.Equal(t, "BackupJobFailed", a.Summary)
		assert.NotContains(t, a.Details, "Query:")
		assert.Equal(t, "Azure Backup", meta["monitoring_service"])
	})

	t.Run("no alertContext at all", func(t *testing.T) {
		body := `{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {"essentials": {"alertId": "a", "alertRule": "Bare", "monitorCondition": "Fired"}}
}`
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)
		assert.Equal(t, "Bare", a.Summary)
	})

	// Neither a condition nor a Prometheus expression: nothing identifies the
	// shape, so essentials alone must still produce a usable alert.
	t.Run("no condition and no expression", func(t *testing.T) {
		body := `{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {"alertId": "p1", "alertRule": "PromRule", "signalType": "Metric",
      "monitoringService": "Prometheus", "monitorCondition": "Fired"},
    "alertContext": {"labels": {"severity": "warning"}, "annotations": {}}
  }
}`
		a, _, _, err := buildAlert([]byte(body))
		require.NoError(t, err)
		assert.Equal(t, "PromRule", a.Summary)
		require.NotNil(t, a.Dedup)
	})

	t.Run("malformed json errors", func(t *testing.T) {
		_, _, _, err := buildAlert([]byte(`{not json`))
		require.Error(t, err)
		assert.NotErrorIs(t, err, errLegacySchema)
	})
}

// Invariants that must hold for every accepted payload, on every branch.
func TestBuildAlert_Invariants(t *testing.T) {
	cases := map[string]string{
		"metric fired":    metricPayload("Fired"),
		"metric resolved": metricPayload("Resolved"),
		"dynamic":         dynamicPayload(),
		"log":             logPayload("Heartbeat", "https://portal.azure.com#@x"),
		"log long":        logPayload(strings.Repeat("q ", 5000), "https://portal.azure.com#@"+strings.Repeat("z", 1000)),
		"service health":  serviceHealthPayload(),
		"prometheus":      prometheusPayload(),
		"webtest":         webtestPayload(),
		"multi resource": strings.Replace(metricPayload("Fired"),
			`"conditionType": "SingleResourceMultipleMetricCriteria"`,
			`"conditionType": "MultipleResourceMultipleMetricCriteria"`, 1),
		"numeric threshold": strings.Replace(metricPayload("Fired"), `"threshold": "0"`, `"threshold": 25`, 1),
		"empty essentials":  `{"schemaId":"azureMonitorCommonAlertSchema","data":{"essentials":{}}}`,
		"empty data":        `{"schemaId":"azureMonitorCommonAlertSchema","data":{}}`,
		"only schema":       `{"schemaId":"azureMonitorCommonAlertSchema"}`,
		"null alertContext": `{"schemaId":"azureMonitorCommonAlertSchema","data":{"essentials":{"alertRule":"x"},"alertContext":null}}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			a, meta, _, err := buildAlert([]byte(body))
			require.NoError(t, err)

			// Empty summary does not error -- it creates a blank, unactionable
			// alert -- so the fallback chain is the only guard.
			assert.NotEmpty(t, a.Summary, "summary must never be empty")
			assert.LessOrEqual(t, len([]rune(a.Summary)), alert.MaxSummaryLength)
			assert.LessOrEqual(t, len([]rune(a.Details)), alert.MaxDetailsLength)

			// A nil dedup would fall back to a content hash that differs between
			// the Fired and Resolved deliveries, breaking the close path.
			require.NotNil(t, a.Dedup, "dedup must never be nil")
			assert.Equal(t, alert.DedupTypeUser, a.Dedup.Type)
			assert.Len(t, a.Dedup.Payload, 64)
			assert.Same(t, a.Dedup, a.DedupKey())

			assert.Equal(t, alert.SourceAzureMonitor, a.Source)

			// Proves the mapper can never produce a client error, and catches a
			// missing SourceAzureMonitor entry in alert.Normalize's OneOf.
			a.ServiceID = testServiceID
			_, err = a.Normalize()
			assert.NoError(t, err)

			total := 0
			for k, v := range meta {
				assert.NotEmpty(t, v, "meta[%s] should have been dropped", k)
				assert.LessOrEqual(t, len([]rune(v)), maxMetaValueLen)
				total += len(k) + len(v)
			}
			assert.Less(t, total, 32*1024)
		})
	}
}

// Distinct alertIds must produce distinct dedup keys, or unrelated alerts would
// collapse onto one.
func TestBuildAlert_DistinctAlertIDs(t *testing.T) {
	first, _, _, err := buildAlert([]byte(metricPayload("Fired")))
	require.NoError(t, err)

	other := strings.Replace(metricPayload("Fired"), "1db044ff-df8f-4064-a559-b9c9f5f4f000", "3a10e1f4-0000-0000-0000-000000000000", 1)
	second, _, _, err := buildAlert([]byte(other))
	require.NoError(t, err)

	require.NotNil(t, first.Dedup)
	require.NotNil(t, second.Dedup)
	assert.NotEqual(t, first.Dedup.Payload, second.Dedup.Payload)
}

func TestStringProps_StripsHTMLAndSkipsNonStrings(t *testing.T) {
	body := `{
  "schemaId": "azureMonitorCommonAlertSchema",
  "data": {
    "essentials": {"alertRule": "x", "monitorCondition": "Fired"},
    "alertContext": {"properties": {
      "title": "<b>Bold</b> title",
      "channels": 1,
      "nested": {"a": "b"},
      "nothing": null,
      "blank": "   "
    }}
  }
}`
	a, _, _, err := buildAlert([]byte(body))
	require.NoError(t, err)

	assert.Contains(t, a.Details, "Title: Bold title")
	assert.NotContains(t, a.Details, "<b>")
	// Non-string values are skipped rather than dumped as raw JSON.
	assert.NotContains(t, a.Details, `{"a"`)
}
