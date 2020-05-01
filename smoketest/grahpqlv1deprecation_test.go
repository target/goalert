package smoketest

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
)

// TestGraphQLV1Deprecation verifies that the /v1/graphql endpoint
// returns a 404 when the API is disabled from the config.
func TestGraphQLV1Deprecation(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, "", "calendar-subscriptions-per-user") // latest migration
	defer h.Close()

	url := h.URL() + "/v1/graphql"

	// ensure api is enabled
	h.SetConfigValue("General.DisableV1GraphQL", "false")

	// test graphql v1 endpoint returns successfully (unauthorized status code)
	resp, _ := http.Get(url)
	assert.Equal(t, resp.StatusCode, 401)

	// disable api
	h.SetConfigValue("General.DisableV1GraphQL", "true")

	// test graphql v1 endpoint returns not found status code
	resp, _ = http.Get(url)
	assert.Equal(t, resp.StatusCode, 404)
}
