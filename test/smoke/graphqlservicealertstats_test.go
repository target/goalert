package smoke

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestGraphQLServiceAlertStats tests the alertStats field for a Service.
// This test validates:
// 1. Correct calculation of averages (avgAckSec, avgCloseSec) from alert_metrics data
// 2. Correct counting of total alerts and escalated alerts
// 3. Proper time window filtering - alerts outside the requested time range are excluded
// 4. Proper data aggregation across multiple time buckets when the time range spans multiple days
//
// The test includes alerts with varying:
// - time_to_ack values (1 hour vs 30 minutes)
// - time_to_close values (2 hours vs 3 hours)
// - escalated status (false vs true)
// - closed_at timestamps (different days to test time filtering)
func TestGraphQLServiceAlertStats(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into escalation_policies (id, name)
		values
			({{uuid "eid"}}, 'esc policy');
		insert into services (id, escalation_policy_id, name)
		values
			({{uuid "sid"}}, {{uuid "eid"}}, 'service');
		insert into alerts (id, service_id, status, summary)
		values
			(1, {{uuid "sid"}}, 'closed', 'alert 1'),
			(2, {{uuid "sid"}}, 'closed', 'alert 2'),
			(3, {{uuid "sid"}}, 'closed', 'alert 3'),
			(4, {{uuid "sid"}}, 'closed', 'alert 4'),
			(5, {{uuid "sid"}}, 'closed', 'alert 5');
		insert into alert_metrics (alert_id, service_id, time_to_ack, time_to_close, escalated, closed_at)
		values
			-- First 4 alerts: consistent timing, all in 2022-01-01
			(1, {{uuid "sid"}}, '1 hour'::interval, '2 hours'::interval, false, '2022-01-01 00:01:00'),
			(2, {{uuid "sid"}}, '1 hour'::interval, '2 hours'::interval, false, '2022-01-01 00:02:00'),
			(3, {{uuid "sid"}}, '1 hour'::interval, '2 hours'::interval, false, '2022-01-01 00:03:00'),
			(4, {{uuid "sid"}}, '1 hour'::interval, '2 hours'::interval, false, '2022-01-01 00:04:00'),
			-- Alert 5: different timing, escalated, closed on 2022-01-02 (outside first window)
			(5, {{uuid "sid"}}, '30 minutes'::interval, '3 hours'::interval, true, '2022-01-02 00:05:00');
	`

	h := harness.NewHarness(t, sql, "om-history-index")
	defer h.Close()

	// Test 1: Query for Jan 1st - should include alerts 1-4 only
	resp1 := h.GraphQLQueryT(t, fmt.Sprintf(`{service(id: "%s") {alertStats(input: {start: "2022-01-01T00:00:00Z", end: "2022-01-02T00:00:00Z"}) {avgAckSec{value}, avgCloseSec{value}, alertCount{value}, escalatedCount{value}}}}`, h.UUID("sid")))

	type bucket struct {
		Start, End time.Time
		Value      float64
	}
	type statsResponse struct {
		Service struct {
			AlertStats struct {
				AvgAckSec      []bucket
				AvgCloseSec    []bucket
				AlertCount     []bucket
				EscalatedCount []bucket
			}
		}
	}
	var result1 statsResponse
	err := json.Unmarshal(resp1.Data, &result1)
	require.NoError(t, err, "should return valid JSON")

	// Validate first query (Jan 1st data - alerts 1-4)
	require.Len(t, result1.Service.AlertStats.AvgAckSec, 1, "should have one time bucket")
	require.Len(t, result1.Service.AlertStats.AvgCloseSec, 1, "should have one time bucket")
	require.Len(t, result1.Service.AlertStats.AlertCount, 1, "should have one time bucket")
	require.Len(t, result1.Service.AlertStats.EscalatedCount, 1, "should have one time bucket")

	// Check averages for first 4 alerts (all 1 hour ack, 2 hour close)
	require.Equal(t, 3600.0, result1.Service.AlertStats.AvgAckSec[0].Value, "average ack time should be 1 hour (3600 sec)")
	require.Equal(t, 7200.0, result1.Service.AlertStats.AvgCloseSec[0].Value, "average close time should be 2 hours (7200 sec)")
	require.Equal(t, 4.0, result1.Service.AlertStats.AlertCount[0].Value, "should have 4 alerts")
	require.Equal(t, 0.0, result1.Service.AlertStats.EscalatedCount[0].Value, "should have 0 escalated alerts")

	// Test 2: Query for Jan 1-2 - should include all 5 alerts
	resp2 := h.GraphQLQueryT(t, fmt.Sprintf(`{service(id: "%s") {alertStats(input: {start: "2022-01-01T00:00:00Z", end: "2022-01-03T00:00:00Z"}) {avgAckSec{value}, avgCloseSec{value}, alertCount{value}, escalatedCount{value}}}}`, h.UUID("sid")))

	var result2 statsResponse
	err = json.Unmarshal(resp2.Data, &result2)
	require.NoError(t, err, "should return valid JSON")

	// Validate second query (Jan 1-2 data - all 5 alerts)
	// We should have 2 time buckets now (one for each day)
	require.Len(t, result2.Service.AlertStats.AvgAckSec, 2, "should have two time buckets")
	require.Len(t, result2.Service.AlertStats.AvgCloseSec, 2, "should have two time buckets")
	require.Len(t, result2.Service.AlertStats.AlertCount, 2, "should have two time buckets")
	require.Len(t, result2.Service.AlertStats.EscalatedCount, 2, "should have two time buckets")

	// First bucket (Jan 1): alerts 1-4
	require.Equal(t, 3600.0, result2.Service.AlertStats.AvgAckSec[0].Value, "Jan 1 average ack time should be 1 hour")
	require.Equal(t, 7200.0, result2.Service.AlertStats.AvgCloseSec[0].Value, "Jan 1 average close time should be 2 hours")
	require.Equal(t, 4.0, result2.Service.AlertStats.AlertCount[0].Value, "Jan 1 should have 4 alerts")
	require.Equal(t, 0.0, result2.Service.AlertStats.EscalatedCount[0].Value, "Jan 1 should have 0 escalated alerts")

	// Second bucket (Jan 2): alert 5 only
	require.Equal(t, 1800.0, result2.Service.AlertStats.AvgAckSec[1].Value, "Jan 2 average ack time should be 30 minutes (1800 sec)")
	require.Equal(t, 10800.0, result2.Service.AlertStats.AvgCloseSec[1].Value, "Jan 2 average close time should be 3 hours (10800 sec)")
	require.Equal(t, 1.0, result2.Service.AlertStats.AlertCount[1].Value, "Jan 2 should have 1 alert")
	require.Equal(t, 1.0, result2.Service.AlertStats.EscalatedCount[1].Value, "Jan 2 should have 1 escalated alert")

	// Test 3: Query for only Jan 2nd - should include only alert 5
	resp3 := h.GraphQLQueryT(t, fmt.Sprintf(`{service(id: "%s") {alertStats(input: {start: "2022-01-02T00:00:00Z", end: "2022-01-03T00:00:00Z"}) {avgAckSec{value}, avgCloseSec{value}, alertCount{value}, escalatedCount{value}}}}`, h.UUID("sid")))

	var result3 statsResponse
	err = json.Unmarshal(resp3.Data, &result3)
	require.NoError(t, err, "should return valid JSON")

	// Validate third query (Jan 2 only - alert 5 only)
	require.Len(t, result3.Service.AlertStats.AvgAckSec, 1, "should have one time bucket")
	require.Equal(t, 1800.0, result3.Service.AlertStats.AvgAckSec[0].Value, "should match alert 5's ack time (30 min)")
	require.Equal(t, 10800.0, result3.Service.AlertStats.AvgCloseSec[0].Value, "should match alert 5's close time (3 hours)")
	require.Equal(t, 1.0, result3.Service.AlertStats.AlertCount[0].Value, "should have 1 alert")
	require.Equal(t, 1.0, result3.Service.AlertStats.EscalatedCount[0].Value, "should have 1 escalated alert")
}
