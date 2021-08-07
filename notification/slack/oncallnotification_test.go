package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestRenderOnCallNotification(t *testing.T) {
	// test with `msg` for the provided user ids
	check := func(desc, expected string, userIDs []string) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			msg := notification.ScheduleOnCallUsers{
				ScheduleName: "schedule.name",
				ScheduleURL:  "schedule.url",
			}

			for _, id := range userIDs {
				msg.Users = append(msg.Users, notification.User{
					ID:   id,
					Name: id + ".name",
					URL:  id + ".url",
				})
			}

			assert.Equal(t, expected, renderOnCallNotificationMessage(msg,

				// define some static mappings for testing graceful fallback
				map[string]string{
					"slack.1": "slack.1.SLACKID",
					"slack.2": "slack.2.SLACKID",
					"slack.3": "slack.3.SLACKID",
				}))
		})
	}

	check("empty", "No users are on-call for <schedule.url|schedule.name>", nil)

	check("fallback", "<foo.url|foo.name> is on-call for <schedule.url|schedule.name>", []string{"foo"})

	check("1 user", "<@slack.1.SLACKID> is on-call for <schedule.url|schedule.name>", []string{"slack.1"})
	check("2 users",
		"<@slack.1.SLACKID> and <@slack.2.SLACKID> are on-call for <schedule.url|schedule.name>",
		[]string{"slack.1", "slack.2"})
	check("3 users",
		"<@slack.1.SLACKID>, <@slack.2.SLACKID>, and <@slack.3.SLACKID> are on-call for <schedule.url|schedule.name>",
		[]string{"slack.1", "slack.2", "slack.3"})

	check("3 users with fallback",
		"<@slack.1.SLACKID>, <foo.url|foo.name>, and <@slack.3.SLACKID> are on-call for <schedule.url|schedule.name>",
		[]string{"slack.1", "foo", "slack.3"})

	t.Run("no panic on nil map", func(t *testing.T) {
		msg := notification.ScheduleOnCallUsers{
			ScheduleName: "schedule.name",
			ScheduleURL:  "schedule.url",
			Users: []notification.User{
				{ID: "foo.SLACKID", Name: "foo.name", URL: "foo.url"},
			},
		}

		assert.Equal(t, "<foo.url|foo.name> is on-call for <schedule.url|schedule.name>", renderOnCallNotificationMessage(msg, nil))
	})

}
