package smoke

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/test/smoke/harness"
)

// TestTwilioSMSNotCMOwner checks that a test sent from a user who is not the
// owner of the contact method returns an error.
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

	cm1 := h.UUID("cm1")

	g := h.GraphQLQuery2(fmt.Sprintf(`
		mutation {
			testContactMethod(id: "%s")
		}
		`, cm1))
	require.Len(t, g.Errors, 1, "errors returned from GraphQL")
	require.Equal(t, "access denied", g.Errors[0].Message)
}
