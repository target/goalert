package smoke

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/test/smoke/harness"
)

// TestUIKDedupClose verifies that a universal integration key's builtin-alert
// action uses an explicit `dedup` param -- instead of summary/details -- to
// match an existing alert for de-duplication and closing.
func TestUIKDedupClose(t *testing.T) {
	t.Parallel()

	const sql = `
		insert into escalation_policies (id, name) values
			({{uuid "eid"}}, 'esc policy');
		insert into services (id, escalation_policy_id, name) values
			({{uuid "sid"}}, {{uuid "eid"}}, 'service');
	`

	h := harness.NewHarnessWithFlags(t, sql, "", expflag.FlagSet{expflag.UnivKeys})
	defer h.Close()

	resp := h.GraphQLQuery2(fmt.Sprintf(`mutation{ createIntegrationKey(input: {name: "key", type: universal, serviceID: "%s"}){ id, href } }`, h.UUID("sid")))
	require.Empty(t, resp.Errors)

	var respData struct {
		CreateIntegrationKey struct {
			ID   uuid.UUID
			Href string
		}
	}
	require.NoError(t, json.Unmarshal(resp.Data, &respData))

	resp = h.GraphQLQuery2(fmt.Sprintf(`
		mutation{
			updateKeyConfig(input: {
				keyID: "%s",
				defaultActions: [
					{dest: {type: "builtin-alert"},
						params: {
							summary: "req.body['summary']",
							dedup: "req.body['dedup']",
							close: "req.body['close']"
						}}
				]
			})
		}`, respData.CreateIntegrationKey.ID))
	require.Empty(t, resp.Errors)

	resp = h.GraphQLQuery2(fmt.Sprintf(`mutation{ generateKeyToken(id: "%s")}`, respData.CreateIntegrationKey.ID))
	require.Empty(t, resp.Errors)
	var gen struct{ GenerateKeyToken string }
	require.NoError(t, json.Unmarshal(resp.Data, &gen))

	post := func(body string) {
		t.Helper()
		req, err := http.NewRequest("POST", respData.CreateIntegrationKey.Href, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+gen.GenerateKeyToken)
		req.Header.Set("Content-Type", "application/json")
		r, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer r.Body.Close()
		data, _ := io.ReadAll(r.Body)
		require.Equalf(t, http.StatusNoContent, r.StatusCode, "response: %s", data)
	}

	type alertNode struct {
		ID      string
		Summary string
		Status  string
	}
	getAlerts := func() []alertNode {
		t.Helper()
		res := h.GraphQLQuery2(fmt.Sprintf(`query{ alerts(input: {filterByServiceID: ["%s"]}){ nodes { id summary status } } }`, h.UUID("sid")))
		require.Empty(t, res.Errors)
		var data struct {
			Alerts struct{ Nodes []alertNode }
		}
		require.NoError(t, json.Unmarshal(res.Data, &data))
		return data.Alerts.Nodes
	}

	// Trigger two distinct alerts, each with its own explicit dedup key.
	post(`{"summary": "first summary", "dedup": "alert-a"}`)
	post(`{"summary": "second alert", "dedup": "alert-b"}`)

	alerts := getAlerts()
	require.Len(t, alerts, 2, "expected two distinct alerts")

	// Re-triggering with the same dedup key but a different summary must be
	// suppressed as a duplicate of the existing alert
	post(`{"summary": "this should be ignored", "dedup": "alert-a"}`)
	alerts = getAlerts()
	require.Len(t, alerts, 2, "duplicate trigger (same dedup key) should not create a new alert")

	var alertA, alertB alertNode
	for _, a := range alerts {
		switch a.Summary {
		case "first summary":
			alertA = a
		case "second alert":
			alertB = a
		}
	}
	require.NotEmpty(t, alertA.ID, "alert A not found")
	require.NotEmpty(t, alertB.ID, "alert B not found")
	assert.Equal(t, "StatusUnacknowledged", alertA.Status)
	assert.Equal(t, "StatusUnacknowledged", alertB.Status)

	// Closing by dedup key "alert-a" must close only that alert, using a
	// different summary than the one it was originally triggered with
	post(`{"summary": "irrelevant close summary", "dedup": "alert-a", "close": true}`)

	alerts = getAlerts()
	require.Len(t, alerts, 2)
	for _, a := range alerts {
		switch a.ID {
		case alertA.ID:
			assert.Equal(t, "StatusClosed", a.Status, "alert A should be closed")
		case alertB.ID:
			assert.Equal(t, "StatusUnacknowledged", a.Status, "alert B should be unaffected")
		}
	}
}
