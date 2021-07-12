package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/notification"
)

func TestRenderOnCallNotification(t *testing.T) {
	slackUserIDs := map[string]string{"slack.1": "slack.1.ID", "slack.2": "slack.2.ID", "slack.3": "slack.3.ID"}
	msg := notification.ScheduleOnCallUsers{
		ScheduleName: "schedule.name",
		ScheduleURL:  "schedule.url",
	}

	check := func(desc, expected string, users []string) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			msg.Users = nil
			for _, u := range users {
				msg.Users = append(msg.Users, notification.User{
					ID:   u,
					Name: u,
					URL:  u + ".url",
				})
			}
			assert.Equal(t, expected, renderOnCallNotificationMessage(msg, slackUserIDs))
		})
	}

	check("empty", "No users are on-call for <schedule.url|schedule.name>", nil)

	check("fallback", "<foo.url|foo> is on-call for <schedule.url|schedule.name>", []string{"foo"})

	check("1 user", "<@slack.1.ID> is on-call for <schedule.url|schedule.name>", []string{"slack.1"})
	check("2 users",
		"<@slack.1.ID> and <@slack.2.ID> are on-call for <schedule.url|schedule.name>",
		[]string{"slack.1", "slack.2"})
	check("3 users",
		"<@slack.1.ID>, <@slack.2.ID>, and <@slack.3.ID> are on-call for <schedule.url|schedule.name>",
		[]string{"slack.1", "slack.2", "slack.3"})

	check("3 users with fallback",
		"<@slack.1.ID>, <foo.url|foo>, and <@slack.3.ID> are on-call for <schedule.url|schedule.name>",
		[]string{"slack.1", "foo", "slack.3"})

	t.Run("no panic on nil map", func(t *testing.T) {
		msg.Users = []notification.User{{ID: "foo.id", Name: "foo", URL: "foo.url"}}

		assert.Equal(t, "<foo.url|foo> is on-call for <schedule.url|schedule.name>", renderOnCallNotificationMessage(msg, nil))
	})

}
