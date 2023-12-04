package smoke

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

func TestGraphQLError(t *testing.T) {
	h := harness.NewHarness(t, "", "ids-to-uuids")
	defer h.Close()

	req, err := http.NewRequest("POST", h.URL()+"/api/graphql", strings.NewReader(`"test"`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.GraphQLToken(harness.DefaultGraphQLAdminUserID))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, 400, resp.StatusCode, "expected 400 error")

	var r struct {
		Errors []struct {
			Message string
		}
	}
	err = json.NewDecoder(resp.Body).Decode(&r)
	require.NoError(t, err)
	require.Len(t, r.Errors, 1)

	require.Equal(t, "json request body could not be decoded: body must be an object, missing '{'", r.Errors[0].Message)
}
