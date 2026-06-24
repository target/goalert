package smoke

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

func TestGraphQLError(t *testing.T) {
	t.Parallel()
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

func TestGraphQLError_TruncatedRequest(t *testing.T) {
	t.Parallel()
	h := harness.NewHarness(t, "", "ids-to-uuids")
	defer h.Close()

	u, err := url.Parse(h.URL() + "/api/graphql")
	require.NoError(t, err)
	c, err := net.Dial("tcp", u.Host)
	require.NoError(t, err)
	defer c.Close()

	_, err = fmt.Fprintf(c, "POST /api/graphql HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nAuthorization: Bearer %s\r\nContent-Length: 400\r\n\r\n{\"test\"", u.Host, h.GraphQLToken(harness.DefaultGraphQLAdminUserID))
	require.NoError(t, err)

	err = c.Close()
	require.NoError(t, err)

	// Test will fail if there is a backend error.
}
