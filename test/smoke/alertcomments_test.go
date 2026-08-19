package smoke

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/test/smoke/harness"
)

// MaxCommentLength mirrors the server-side limit.
const MaxCommentLength = alert.MaxCommentLength

const commentsInitSQL = `
	insert into users (id, name, email, role)
	values
		({{uuid "user1"}}, 'bob', 'bob@example.com', 'user'),
		({{uuid "user2"}}, 'ann', 'ann@example.com', 'user');

	insert into escalation_policies (id, name)
	values ({{uuid "eid"}}, 'esc policy');

	insert into services (id, escalation_policy_id, name)
	values ({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	-- open alerts must carry a dedup_key (dedup_key_only_for_open_alerts)
	insert into alerts (id, service_id, summary, status, dedup_key)
	values (1, {{uuid "sid"}}, 'testing', 'triggered', 'auto:1:foo');`

type commentsResult struct {
	Alert struct {
		Comments []struct {
			ID   string
			Body string
			User *struct {
				ID   string
				Name string
			}
			CreatedAt string
		}
	}
}

func queryComments(t *testing.T, h *harness.Harness, alertID int) commentsResult {
	t.Helper()

	g := h.GraphQLQueryUserT(t, h.UUID("user1"), fmt.Sprintf(`
		query {
			alert(id: %d) {
				comments { id body createdAt user { id name } }
			}
		}
	`, alertID))
	for _, err := range g.Errors {
		t.Fatal("GraphQL Error:", err.Message)
	}

	var res commentsResult
	require.NoError(t, json.Unmarshal(g.Data, &res))
	return res
}

func addComment(t *testing.T, h *harness.Harness, asUser string, alertID int, body string) string {
	t.Helper()

	g := h.GraphQLQueryUserT(t, asUser, fmt.Sprintf(`
		mutation {
			addAlertComment(input: { alertID: %d, body: %q }) { id }
		}
	`, alertID, body))
	for _, err := range g.Errors {
		t.Fatal("GraphQL Error:", err.Message)
	}

	var res struct {
		AddAlertComment struct{ ID string }
	}
	require.NoError(t, json.Unmarshal(g.Data, &res))
	return res.AddAlertComment.ID
}

// countComments reads the table directly, so cascade behavior is verified
// against the database rather than merely being unreachable through the API.
func countComments(t *testing.T, h *harness.Harness) int {
	t.Helper()

	var n int
	err := h.App().DB().QueryRowContext(context.Background(),
		"select count(*) from alert_comments").Scan(&n)
	require.NoError(t, err)
	return n
}

// TestAlertComments verifies that comments record who, when and what, and that
// any user may comment on any alert.
func TestAlertComments(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, commentsInitSQL, "add-alert-comments")
	defer h.Close()

	before := time.Now()

	addComment(t, h, h.UUID("user1"), 1, "looking into this")
	// user2 is on no escalation policy for this service -- anyone can comment.
	addComment(t, h, h.UUID("user2"), 1, "handing off, out of ideas")

	res := queryComments(t, h, 1)
	require.Len(t, res.Alert.Comments, 2, "expected both comments")

	// oldest first
	first, second := res.Alert.Comments[0], res.Alert.Comments[1]

	assert.Equal(t, "looking into this", first.Body)
	require.NotNil(t, first.User, "expected an author")
	assert.Equal(t, h.UUID("user1"), first.User.ID)
	assert.Equal(t, "bob", first.User.Name)

	assert.Equal(t, "handing off, out of ideas", second.Body)
	require.NotNil(t, second.User)
	assert.Equal(t, h.UUID("user2"), second.User.ID)

	createdAt, err := time.Parse(time.RFC3339, first.CreatedAt)
	require.NoError(t, err, "createdAt should be a timestamp")
	assert.False(t, createdAt.Before(before.Add(-time.Minute)),
		"createdAt should record when the comment was made")
}

// TestAlertCommentsValidation verifies that empty and oversized comments are
// rejected, and that the body is trimmed.
func TestAlertCommentsValidation(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, commentsInitSQL, "add-alert-comments")
	defer h.Close()

	expectErr := func(body string) {
		t.Helper()
		g := h.GraphQLQueryUserT(t, h.UUID("user1"), fmt.Sprintf(`
			mutation {
				addAlertComment(input: { alertID: 1, body: %q }) { id }
			}
		`, body))
		assert.NotEmpty(t, g.Errors, "expected rejection for body %q", body)
	}

	expectErr("")
	expectErr("   ")
	expectErr(strings.Repeat("a", MaxCommentLength+1))

	// surrounding whitespace is trimmed rather than rejected
	addComment(t, h, h.UUID("user1"), 1, "  spaced out  ")
	res := queryComments(t, h, 1)
	require.Len(t, res.Alert.Comments, 1)
	assert.Equal(t, "spaced out", res.Alert.Comments[0].Body)
}

// TestAlertCommentsDelete verifies that users can delete their own comments but
// not other people's.
func TestAlertCommentsDelete(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, commentsInitSQL, "add-alert-comments")
	defer h.Close()

	id := addComment(t, h, h.UUID("user1"), 1, "mine to remove")

	// user2 must not be able to delete user1's comment
	g := h.GraphQLQueryUserT(t, h.UUID("user2"), fmt.Sprintf(`
		mutation { deleteAlertComment(id: %q) }
	`, id))
	assert.NotEmpty(t, g.Errors, "another user must not delete someone else's comment")
	assert.Len(t, queryComments(t, h, 1).Alert.Comments, 1, "comment should survive")

	// the author can
	g = h.GraphQLQueryUserT(t, h.UUID("user1"), fmt.Sprintf(`
		mutation { deleteAlertComment(id: %q) }
	`, id))
	for _, err := range g.Errors {
		t.Fatal("GraphQL Error:", err.Message)
	}
	assert.Empty(t, queryComments(t, h, 1).Alert.Comments, "comment should be gone")
}

// TestAlertCommentsCleanupCascade is the guarantee against data sprawl: when an
// alert is purged by the cleanup job, its comments must go with it.
//
// This asserts against the alert_comments table directly, because once the
// alert is deleted the comments are unreachable through the API whether or not
// they were actually removed.
func TestAlertCommentsCleanupCascade(t *testing.T) {
	t.Parallel()

	sql := `
	insert into users (id, name, email, role)
	values ({{uuid "user1"}}, 'bob', 'bob@example.com', 'user');

	insert into escalation_policies (id, name)
	values ({{uuid "eid"}}, 'esc policy');

	insert into services (id, escalation_policy_id, name)
	values ({{uuid "sid"}}, {{uuid "eid"}}, 'service');

	insert into alerts (id, service_id, summary, status, created_at)
	values
		(1, {{uuid "sid"}}, 'recent', 'closed', now()),
		(2, {{uuid "sid"}}, 'stale', 'closed', now() - '2 days'::interval);`

	h := harness.NewHarness(t, sql, "add-alert-comments")
	defer h.Close()

	addComment(t, h, h.UUID("user1"), 1, "on the alert that stays")
	addComment(t, h, h.UUID("user1"), 2, "on the alert that gets purged")
	addComment(t, h, h.UUID("user1"), 2, "and another on the purged one")

	require.Equal(t, 3, countComments(t, h), "expected all three comments")

	cfg := h.Config()
	cfg.Maintenance.AlertCleanupDays = 1
	h.RestartGoAlertWithConfig(cfg)

	assert.EventuallyWithT(t, func(t *assert.CollectT) {
		var data struct {
			A *struct{ ID string }
		}
		res := h.GraphQLQuery2("{a:alert(id: 2){id}}")
		assert.Empty(t, res.Errors)
		assert.NoError(t, json.Unmarshal(res.Data, &data))
		assert.Nil(t, data.A, "stale alert should be cleaned up")
	}, 15*time.Second, time.Second)

	// Only the surviving alert's comment remains -- the two belonging to the
	// purged alert were removed by the FK cascade, with no cleanup job of
	// their own.
	assert.Equal(t, 1, countComments(t, h),
		"comments on the purged alert should be cascade-deleted")

	remaining := queryComments(t, h, 1)
	require.Len(t, remaining.Alert.Comments, 1)
	assert.Equal(t, "on the alert that stays", remaining.Alert.Comments[0].Body)
}

// TestAlertCommentsUserDeleted verifies the other half of the retention story:
// deleting a user must not erase their comments, only the attribution.
func TestAlertCommentsUserDeleted(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, commentsInitSQL, "add-alert-comments")
	defer h.Close()

	addComment(t, h, h.UUID("user2"), 1, "survives its author")

	// deleting a user requires admin, so run this as the harness admin
	g := h.GraphQLQuery2(fmt.Sprintf(`
		mutation { deleteAll(input: [{ type: user, id: "%s" }]) }
	`, h.UUID("user2")))
	for _, err := range g.Errors {
		t.Fatal("GraphQL Error:", err.Message)
	}

	res := queryComments(t, h, 1)
	require.Len(t, res.Alert.Comments, 1, "comment must outlive its author")
	assert.Equal(t, "survives its author", res.Alert.Comments[0].Body)
	assert.Nil(t, res.Alert.Comments[0].User, "attribution is dropped, not the comment")
}
