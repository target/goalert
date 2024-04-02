package smoke

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestAlertMetadata tests that creating an alert with metadata results in the metadata being included in the alert.
//
// - Create with GraphQL
// - Create with Generic API
// - Verify metadata is included in alert from GraphQL
func TestAlertMetadata(t *testing.T) {
	const sql = `
	insert into escalation_policies (id, name) 
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name) 
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');
	insert into integration_keys (id, type, name, service_id)
	values
		({{uuid "int_key"}}, 'generic', 'my key', {{uuid "sid"}});
	`

	h := harness.NewHarness(t, sql, "")
	defer h.Close()

	h.GraphQLQuery2(`mutation{createAlert(input:{serviceID:"` + h.UUID("sid") + `",summary:"gql",meta:[{key:"gql", value: "gqlvalue"}]}){id}}`)

	resp, err := http.Post(h.URL()+"/api/v2/generic/incoming?summary=gen_form&meta=form1key=form1value&meta=form2key=form2value&token="+h.UUID("int_key"), "", nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	var bod struct {
		Summary string            `json:"summary"`
		Meta    map[string]string `json:"meta"`
	}
	bod.Summary = "gen_json"
	bod.Meta = map[string]string{"jsonkey": "jsonvalue"}
	data, err := json.Marshal(bod)
	require.NoError(t, err)
	resp, err = http.Post(h.URL()+"/api/v2/generic/incoming?token="+h.UUID("int_key"), "application/json", bytes.NewReader(data))
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	res := h.GraphQLQuery2(`query{alerts{nodes{alertID summary meta{key value}}}}`)

	var result struct {
		Alerts struct {
			Nodes []struct {
				AlertID int
				Summary string
				Meta    []struct {
					Key   string
					Value string
				}
			}
		}
	}
	require.NoError(t, json.Unmarshal(res.Data, &result), "failed to parse response: %s", string(res.Data))

	sort.Slice(result.Alerts.Nodes, func(i, j int) bool { return result.Alerts.Nodes[i].AlertID < result.Alerts.Nodes[j].AlertID })
	require.Len(t, result.Alerts.Nodes, 3)

	require.Equal(t, "gql", result.Alerts.Nodes[0].Summary)
	require.Len(t, result.Alerts.Nodes[0].Meta, 1)
	require.Equal(t, "gql", result.Alerts.Nodes[0].Meta[0].Key)
	require.Equal(t, "gqlvalue", result.Alerts.Nodes[0].Meta[0].Value)

	require.Equal(t, "gen_form", result.Alerts.Nodes[1].Summary)
	require.Len(t, result.Alerts.Nodes[1].Meta, 2)
	require.Equal(t, "form1key", result.Alerts.Nodes[1].Meta[0].Key)
	require.Equal(t, "form1value", result.Alerts.Nodes[1].Meta[0].Value)
	require.Equal(t, "form2key", result.Alerts.Nodes[1].Meta[1].Key)
	require.Equal(t, "form2value", result.Alerts.Nodes[1].Meta[1].Value)

	require.Equal(t, "gen_json", result.Alerts.Nodes[2].Summary)
	require.Len(t, result.Alerts.Nodes[2].Meta, 1)
	require.Equal(t, "jsonkey", result.Alerts.Nodes[2].Meta[0].Key)
	require.Equal(t, "jsonvalue", result.Alerts.Nodes[2].Meta[0].Value)

}
