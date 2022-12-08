package smoke

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioSMS checks that a test SMS is processed.
func TestTwilioSMSNotCMOwner(t *testing.T) {
	t.Parallel()

	sqlQuery := `
	insert into users (id, name, email) 
	values 
		({{uuid "user"}}, 'bob', 'joe');
	insert into user_contact_methods (id, user_id, name, type, value) 
	values
	    ({{uuid "cm1"}}, {{uuid "user"}}, 'personal', 'SMS', {{phone "1"}});
`
	h := harness.NewHarness(t, sqlQuery, "add-verification-code")
	defer h.Close()

	doQL := func(query string) {
		g := h.GraphQLQuery2(query)
		require.Len(t, g.Errors, 1, "errors returned from GraphQL")
		require.Equal(t, "contact method does not belong to user", g.Errors[0].Message)
	}
	cm1 := h.UUID("cm1")

	doQL(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
		`, cm1))
}
