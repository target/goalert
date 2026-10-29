package smoke

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestGraphQLRestrictResourceChanges tests that General.RestrictResourceChanges
// limits changes to shared resources to admins, while users can still manage alerts
// and schedule overrides.
func TestGraphQLRestrictResourceChanges(t *testing.T) {
	t.Parallel()

	const sql = `
	insert into users (id, name, email, role)
	values
		({{uuid "user"}}, 'bob', 'bob@example.com', 'user');
	insert into escalation_policies (id, name)
	values
		({{uuid "eid"}}, 'esc policy');
	insert into services (id, escalation_policy_id, name)
	values
		({{uuid "sid"}}, {{uuid "eid"}}, 'service');
	insert into integration_keys (id, name, type, service_id)
	values
		({{uuid "ikey"}}, 'key', 'generic', {{uuid "sid"}});
	insert into rotations (id, name, description, type, start_time, time_zone)
	values
		({{uuid "rid"}}, 'rotation', 'test', 'daily', now(), 'UTC');
	insert into schedules (id, name, time_zone)
	values
		({{uuid "schedID"}}, 'schedule', 'UTC');
	insert into alerts (id, service_id, summary, status, dedup_key)
	values
		(1, {{uuid "sid"}}, 'testing', 'triggered', 'test:1:foo');
`

	h := harness.NewHarness(t, sql, "")
	defer h.Close()

	userID := h.UUID("user")

	requireOK := func(t *testing.T, userID, query string) {
		t.Helper()
		resp := h.GraphQLQueryUserT(t, userID, query)
		require.Empty(t, resp.Errors, "expected query to succeed")
	}
	requireDenied := func(t *testing.T, userID, query string) {
		t.Helper()
		resp := h.GraphQLQueryUserT(t, userID, query)
		require.NotEmpty(t, resp.Errors, "expected query to be denied")
		require.Contains(t, resp.Errors[0].Message, "access denied")
	}

	updateService := func(name string) string {
		return fmt.Sprintf(`mutation{updateService(input:{id: "%s", name: "%s"})}`, h.UUID("sid"), name)
	}
	listKeys := fmt.Sprintf(`query{service(id: "%s"){integrationKeys{id}}}`, h.UUID("sid"))

	requireOK(t, userID, updateService("user-unrestricted"))
	requireOK(t, userID, listKeys)

	h.SetConfigValue("General.RestrictResourceChanges", "true")

	requireDenied(t, userID, updateService("user-restricted"))
	requireDenied(t, userID, listKeys)
	requireDenied(t, userID, fmt.Sprintf(`mutation{createIntegrationKey(input:{serviceID: "%s", type: generic, name: "new key"}){id}}`, h.UUID("sid")))
	requireDenied(t, userID, fmt.Sprintf(`mutation{updateEscalationPolicy(input:{id: "%s", name: "renamed"})}`, h.UUID("eid")))
	requireDenied(t, userID, fmt.Sprintf(`mutation{updateRotation(input:{id: "%s", name: "renamed"})}`, h.UUID("rid")))

	requireOK(t, userID, `mutation{updateAlerts(input:{alertIDs: [1], newStatus: StatusClosed}){id}}`)

	start := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	end := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	requireOK(t, userID, fmt.Sprintf(`mutation{createUserOverride(input:{scheduleID: "%s", addUserID: "%s", start: "%s", end: "%s"}){id}}`, h.UUID("schedID"), userID, start, end))

	requireOK(t, harness.DefaultGraphQLAdminUserID, updateService("admin-restricted"))
	requireOK(t, harness.DefaultGraphQLAdminUserID, listKeys)
}
