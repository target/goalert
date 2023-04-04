package slack

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util/log"
)

type UserGroupSender struct {
	*ChannelSender
}

var _ notification.Sender = (*UserGroupSender)(nil)

func (s *ChannelSender) UserGroupSender() *UserGroupSender {
	return &UserGroupSender{s}
}

func (s *UserGroupSender) Send(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return nil, err
	}

	t, ok := msg.(*notification.ScheduleOnCallUsers)
	if !ok {
		return nil, errors.Errorf("unsupported message type: %T", msg)
	}

	if t.Dest.Type != notification.DestTypeSlackUG {
		return nil, errors.Errorf("unsupported destination type: %s", t.Dest.Type.String())
	}

	teamID, err := s.TeamID(ctx)
	if err != nil {
		return nil, fmt.Errorf("lookup team ID: %w", err)
	}

	var userIDs []string
	for _, u := range t.Users {
		userIDs = append(userIDs, u.ID)
	}

	userSlackIDs := make(map[string]string, len(t.Users))
	err = s.cfg.UserStore.AuthSubjectsFunc(ctx, "slack:"+teamID, userIDs, func(sub user.AuthSubject) error {
		userSlackIDs[sub.UserID] = sub.SubjectID
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("lookup user slack IDs: %w", err)
	}

	var slackUsers []string
	var missing []notification.User
	for _, u := range t.Users {
		slackID, ok := userSlackIDs[u.ID]
		if !ok {
			missing = append(missing, u)
			continue
		}

		slackUsers = append(slackUsers, slackID)
	}

	ugID, chanID, _ := strings.Cut(t.Dest.Value, ":")

	if len(missing) > 0 {
		// TODO: add link action button to invite missing users
		var buf bytes.Buffer
		err := userGroupErrorMissing.Execute(&buf, userGroupError{
			GroupID: ugID,
			Missing: missing,
			Linked:  slackUsers,
		})
		if err != nil {
			return nil, fmt.Errorf("execute template: %w", err)
		}
		err = s.withClient(ctx, func(c *slack.Client) error {
			_, _, err := c.PostMessageContext(ctx, chanID, slack.MsgOptionText(buf.String(), false))
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("send message: %w", err)
		}
		return &notification.SentMessage{State: notification.StateSent, StateDetails: "missing users, sent error to channel"}, nil
	}

	if len(slackUsers) == 0 {
		var buf bytes.Buffer
		err := userGroupErrorEmpty.Execute(&buf, userGroupError{
			GroupID:      ugID,
			ScheduleID:   t.ScheduleID,
			ScheduleName: t.ScheduleName,
		})
		if err != nil {
			return nil, fmt.Errorf("execute template: %w", err)
		}
		err = s.withClient(ctx, func(c *slack.Client) error {
			_, _, err := c.PostMessageContext(ctx, chanID, slack.MsgOptionText(buf.String(), false))
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("send message: %w", err)
		}
		return &notification.SentMessage{State: notification.StateSent, StateDetails: "empty user-group, sent error to channel"}, nil
	}

	err = s.withClient(ctx, func(c *slack.Client) error {
		_, err := c.UpdateUserGroupMembersContext(ctx, ugID, strings.Join(slackUsers, ","))
		return err
	})
	if err != nil {
		errID := uuid.New()
		log.Log(log.WithField(ctx, "SlackUGErrorID", errID), err)
		var buf bytes.Buffer
		err := userGroupErrorUpdate.Execute(&buf, userGroupError{
			ErrorID: errID,
			GroupID: ugID,
		})
		if err != nil {
			return nil, fmt.Errorf("execute template: %w", err)
		}
		err = s.withClient(ctx, func(c *slack.Client) error {
			_, _, err := c.PostMessageContext(ctx, chanID, slack.MsgOptionText(buf.String(), false))
			return err
		})
		if err != nil {
			return nil, fmt.Errorf("send message: %w", err)
		}
		return &notification.SentMessage{State: notification.StateSent, StateDetails: "failed to update user-group, sent error to channel and log"}, nil
	}

	return &notification.SentMessage{State: notification.StateDelivered}, nil
}

type userGroupError struct {
	ErrorID      uuid.UUID
	GroupID      string
	ScheduleID   string
	ScheduleName string
	Missing      []notification.User
	Linked       []string

	callbackFunc func(string, ...url.Values) string
}

func (e userGroupError) ErrorRef() string {
	return e.ErrorID.String()
}

func (e userGroupError) GroupRef() string {
	return fmt.Sprintf("<@%s>", e.GroupID)
}

func (e userGroupError) MissingUserRefs() string {
	var refs []string
	for _, u := range e.Missing {
		urlStr := e.callbackFunc(fmt.Sprintf("users/%s", url.PathEscape(u.ID)))
		refs = append(refs, slackLink(urlStr, u.Name))
	}
	return strings.Join(refs, ", ")
}

func (e userGroupError) LinkedUserRefs() string {
	var refs []string
	for _, u := range e.Linked {
		refs = append(refs, fmt.Sprintf("<@%s>", u))
	}
	return strings.Join(refs, " ")
}

func (e userGroupError) ScheduleRef() string {
	urlStr := e.callbackFunc("schedules/" + url.PathEscape(e.ScheduleID))
	return slackLink(urlStr, e.ScheduleName)
}

func slackLink(url, label string) string {
	return fmt.Sprintf("<%s|%s>", slackutilsx.EscapeMessage(url), slackutilsx.EscapeMessage(label))
}

var userGroupErrorMissing = template.Must(template.New("userGroupErrorMissing").Parse(`Hey everyone! I couldn't update {{.GroupRef}} because I couldn't find the following user(s) in Slack: {{.MissingUserRefs}}

If you could have them click the button below to connect their Slack account, that would be great! Hopefully I'll be able to update the user-group next time.

{{.LinkedUserRefs}}`))

var userGroupErrorEmpty = template.Must(template.New("userGroupErrorMissing").Parse(`Hey everyone! I couldn't update {{.GroupRef}} because there is nobody on-call for {{.ScheduleRef}}.

Since a Slack user-group cannot be empty, I'm going to leave it as-is for now.`))

var userGroupErrorUpdate = template.Must(template.New("userGroupErrorMissing").Parse(`Hey everyone! I couldn't update {{.GroupRef}} because I ran into a problem. Maybe touch base with the GoAlert admin(s) to see if they can help? I'm sorry for the inconvenience!

Here's the ID I left with the error in my logs so they can find it: {{.ErrorRef}}`))
