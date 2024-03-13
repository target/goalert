package smoke

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/test/smoke/harness"
)

// TestGQLAPIKeys tests most operations on API keys API via GraphQL endpoint.
func TestGQLAPIKeys(t *testing.T) {
	const keyQueryDoc = `
	query ListAPIKeys {
		gqlAPIKeys {
			id
			name
		}
	}
	
	mutation DeleteAPIKey($id: ID!) {
		deleteGQLAPIKey(id: $id)
	}
	
	query ServiceInfo($firstID: ID!) {
	  service(id: $firstID) {
		id
	  }
	}
	
	query ServiceInfo2($secondID: ID!) {
	  service(id: $secondID) {
		id
	  }
	}
	`

	h := harness.NewHarnessWithFlags(t, "", "", expflag.FlagSet{expflag.GQLAPIKey})
	defer h.Close()

	var createVars struct {
		Expires string `json:"expires,omitempty"`
		Query   string `json:"query,omitempty"`
	}
	createVars.Expires = time.Now().Add(time.Hour).Format(time.RFC3339)
	createVars.Query = keyQueryDoc
	// create API key
	gqlResp := h.GraphQLQueryUserVarsT(t, harness.DefaultGraphQLAdminUserID, `
	mutation($expires: ISOTimestamp!, $query: String!){
		createGQLAPIKey(input:{
			name:"test",
			description:"desc",
			expiresAt: $expires,
			role: admin,
			query: $query
		}) {id, token}
	}`, "", createVars)
	var keyResp struct {
		CreateGQLAPIKey struct {
			ID    string
			Token string
		}
	}
	err := json.Unmarshal(gqlResp.Data, &keyResp)
	require.NoError(t, err)
	require.Empty(t, gqlResp.Errors)

	var reqData struct {
		Op string `json:"operationName"`
		V  struct {
			ID     string `json:"id,omitempty"`
			First  string `json:"firstID,omitempty"`
			Second string `json:"secondID,omitempty"`
		} `json:"variables"`
	}
	reqData.Op = "ServiceInfo"
	reqData.V.First = "00000000-0000-0000-0000-000000000001"
	data, err := json.Marshal(reqData)
	require.NoError(t, err)

	t.Log("Data:", string(data))

	req, err := http.NewRequest("POST", h.URL()+"/api/graphql", bytes.NewReader(data))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+keyResp.CreateGQLAPIKey.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var respData struct {
		Errors []struct{ Message string }
	}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	require.NoError(t, err)
	require.Empty(t, respData.Errors)

}
