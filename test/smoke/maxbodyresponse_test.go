package smoke

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

func TestMaxBodyResponse(t *testing.T) {
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
		({{uuid "int_key"}}, 'generic', 'my key', {{uuid "sid"}});
	`

	h := harness.NewHarness(t, sql, "add-generic-integration-key")
	defer h.Close()

	u := h.URL() + "/v1/api/alerts?key=" + h.UUID("int_key")

	// Create a 1MB payload (1024 * 1024 bytes = 1MB)
	// This exceeds the default 256KB limit
	largePayload := strings.Repeat("x", 1024*1024)
	payload := `{"summary": "large payload", "details": "` + largePayload + `"}`

	resp, err := http.Post(u, "application/json", bytes.NewBufferString(payload))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should get 413 Request Entity Too Large
	// Let's first check what we actually got and the response body for debugging
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Logf("Status Code: %d, Body: %s", resp.StatusCode, string(body))

	require.Equal(t, 413, resp.StatusCode, "expected 413 Request Entity Too Large")

	bodyStr := string(body)
	require.Contains(t, bodyStr, "Request Entity Too Large")
	require.Contains(t, bodyStr, "max body:")
}
