package slack

import (
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/slack-go/slack/slackutilsx"
	"github.com/target/goalert/notification"
)

// userGroupError has information about an error that occurred while updating a user-group.
type userGroupError struct {
	ErrorID      uuid.UUID
	GroupID      string
	ScheduleID   string
	ScheduleName string
	Missing      []notification.User

	callbackFunc func(string, ...url.Values) string
}

// ErrorRef returns a string that can be used to reference the error in a Slack message.
func (e userGroupError) ErrorRef() string {
	return fmt.Sprintf("`SlackUGErrorID=%s`", e.ErrorID.String())
}

// GroupRef returns a string that can be used to reference the user-group in a Slack message.
func (e userGroupError) GroupRef() string {
	return fmt.Sprintf("<!subteam^%s>", e.GroupID)
}

// MissingUserRefs returns a string that can be used to reference the missing users in a Slack message.
func (e userGroupError) MissingUserRefs() string {
	var refs []string
	for _, u := range e.Missing {
		urlStr := e.callbackFunc(fmt.Sprintf("users/%s", url.PathEscape(u.ID)))
		refs = append(refs, slackLink(urlStr, u.Name))
	}
	return strings.Join(refs, ", ")
}

// ScheduleRef returns a string that can be used to reference the schedule in a Slack message.
func (e userGroupError) ScheduleRef() string {
	urlStr := e.callbackFunc(fmt.Sprintf("schedules/%s", url.PathEscape(e.ScheduleID)))
	return slackLink(urlStr, e.ScheduleName)
}

// slackLink returns a string that can be used to reference a URL in a Slack message.
func slackLink(url, label string) string {
	return fmt.Sprintf("<%s|%s>", slackutilsx.EscapeMessage(url), slackutilsx.EscapeMessage(label))
}

// userGroupErrorMissing is a template for when a user-group update fails because one or more users are missing from Slack.
var userGroupErrorMissing = template.Must(template.New("userGroupErrorMissing").Parse(`Hey everyone! I couldn't update {{.GroupRef}} because I couldn't find the following user(s) in Slack: {{.MissingUserRefs}}

If you could have them add a SLACK_DM contact method from their respective GoAlert profile page(s), that would be great! Hopefully I'll be able to update the user-group next time.`))

// userGroupErrorEmpty is a template for when a user-group update fails because there are no users on-call.
var userGroupErrorEmpty = template.Must(template.New("userGroupErrorEmpty").Parse(`Hey everyone! I couldn't update {{.GroupRef}} because there is nobody on-call for {{.ScheduleRef}}.

Since a Slack user-group cannot be empty, I'm going to leave it as-is for now.`))

// userGroupErrorUpdate is a template for when a user-group update fails for any reason other than missing users.
var userGroupErrorUpdate = template.Must(template.New("userGroupErrorUpdate").Parse(`Hey everyone! I couldn't update {{.GroupRef}} because I ran into a problem. Maybe touch base with the GoAlert admin(s) to see if they can help? I'm sorry for the inconvenience!

Here's the ID I left with the error in my logs so they can find it:
{{.ErrorRef}}`))
