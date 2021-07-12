package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/target/goalert/notification"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util/log"
)

// onCallNotificationText will return text intended to be sent to Slack representing a ScheduleOnCallUsers notification.
//
// It gracefully degrades to excluding slack IDs when there is an error fetching the required information (e.g., team ID or
// auth subjects).
func (s *ChannelSender) onCallNotificationText(ctx context.Context, t notification.ScheduleOnCallUsers) string {
	if len(t.Users) == 0 {
		return renderOnCallNotificationMessage(t, nil)
	}

	teamID, err := s.TeamID(ctx)
	if err != nil {
		log.Log(ctx, fmt.Errorf("lookup team ID: %w", err))
		return renderOnCallNotificationMessage(t, nil)
	}

	userIDs := make([]string, len(t.Users))
	for i, u := range t.Users {
		userIDs[i] = u.ID
	}

	userSlackIDs := make(map[string]string, len(t.Users))
	err = s.cfg.UserStore.AuthSubjectsFunc(ctx, "slack:"+teamID, func(sub user.AuthSubject) error {
		userSlackIDs[sub.UserID] = sub.SubjectID
		return nil
	}, userIDs...)
	if err != nil {
		log.Log(ctx, fmt.Errorf("lookup auth subjects for slack: %w", err))
		// handled error by logging, continue on to render message with any included slack IDs
	}

	return renderOnCallNotificationMessage(t, userSlackIDs)
}

// renderOnCallNotificationMessage will render a message for Slack including links for the schedule and any users.
//
// If a user's ID is available in userSlackIDs, an `@` user mention will be used in place of a link to the GoAlert user's detail page.
func renderOnCallNotificationMessage(msg notification.ScheduleOnCallUsers, userSlackIDs map[string]string) string {
	suffix := fmt.Sprintf("on-call for <%s|%s>", msg.ScheduleURL, msg.ScheduleName)

	var userLinks []string
	for _, u := range msg.Users {
		var subjectID string
		if userSlackIDs != nil {
			subjectID = userSlackIDs[u.ID]
		}
		if subjectID == "" {
			// fallback to a link to the GoAlert user
			userLinks = append(userLinks, fmt.Sprintf("<%s|%s>", u.URL, u.Name))
			continue
		}

		userLinks = append(userLinks, fmt.Sprintf("<@%s>", subjectID))
	}

	if len(userLinks) == 0 {
		return "No users are " + suffix
	}
	if len(userLinks) == 1 {
		return fmt.Sprintf("%s is %s", userLinks[0], suffix)
	}
	if len(userLinks) == 2 {
		return fmt.Sprintf("%s and %s are %s", userLinks[0], userLinks[1], suffix)
	}

	return fmt.Sprintf("%s, and %s are %s", strings.Join(userLinks[:len(userLinks)-1], ", "), userLinks[len(userLinks)-1], suffix)
}
