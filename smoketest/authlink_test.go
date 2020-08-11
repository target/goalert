package smoketest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/smoketest/harness"
)

// TestAuthLink ensures the "happy path" is functional for linking a mobile device.
func TestAuthLink(t *testing.T) {
	h := harness.NewHarness(t, "", "")
	defer h.Close()

	gResp := h.GraphQLQuery2(`mutation{createAuthLink{id, claimCode}}`)
	require.Empty(t, gResp.Errors)

	var claim struct {
		CreateAuthLink struct {
			ID        string
			ClaimCode string
		}
	}
	err := json.Unmarshal(gResp.Data, &claim)
	require.NoError(t, err)

	v := make(url.Values)
	v.Set("code", claim.CreateAuthLink.ClaimCode)
	resp, err := http.PostForm(h.URL()+"/api/v2/identity/providers/link", v)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var token struct {
		VerifyCode string
		AuthToken  string
	}
	err = json.Unmarshal(data, &token)
	require.NoError(t, err)

	gResp = h.GraphQLQuery2(fmt.Sprintf(`mutation{verifyAuthLink(input:{id: "%s", code: "%s"})}`, claim.CreateAuthLink.ID, token.VerifyCode))
	require.Empty(t, gResp.Errors)

	var verify struct {
		VerifyAuthLink bool
	}
	err = json.Unmarshal(gResp.Data, &verify)
	require.NoError(t, err)
	require.True(t, verify.VerifyAuthLink)

	v = make(url.Values)
	v.Set("AuthToken", token.AuthToken)
	resp, err = http.PostForm(h.URL()+"/api/v2/identity/providers/link/auth", v)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	data, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var query struct {
		Query string
	}
	query.Query = "{user{id}}"
	queryData, err := json.Marshal(query)
	require.NoError(t, err)
	resp, err = http.Post(h.URL()+"/api/graphql?token="+string(data), "application/json", bytes.NewReader(queryData))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	var check struct {
		Data struct{ User struct{ ID string } }
	}

	data, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(data, &check)
	require.NoError(t, err)

	require.Equal(t, harness.DefaultGraphQLAdminUserID, check.Data.User.ID)
}
