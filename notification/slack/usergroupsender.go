package slack

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util/log"
)

// UserGroupSender processes on-call notifications by updating the members of a Slack user group.
type UserGroupSender struct {
	*ChannelSender
}

var _ notification.Sender = (*UserGroupSender)(nil)

// UserGroupSender returns a new UserGroupSender wrapping the given ChannelSender.
func (s *ChannelSender) UserGroupSender() *UserGroupSender {
	return &UserGroupSender{s}
}

// Send implements notification.Sender.
func (s *UserGroupSender) Send(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return nil, err
	}

	t, ok := msg.(notification.ScheduleOnCallUsers)
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
	cfg := config.FromContext(ctx)

	// If any users are missing, we need to abort and let the channel know.
	if len(missing) > 0 {
		// TODO: add link action button to invite missing users
		var buf bytes.Buffer
		err := userGroupErrorMissing.Execute(&buf, userGroupError{
			GroupID:      ugID,
			Missing:      missing,
			Linked:       slackUsers,
			callbackFunc: cfg.CallbackURL,
		})
		if err != nil {
			return nil, fmt.Errorf("execute template: %w", err)
		}
		err = s.withClient(ctx, func(c *slack.Client) error {
			_, _, err := c.PostMessageContext(ctx, chanID, slack.MsgOptionText(buf.String(), false))
			if err != nil {
				return fmt.Errorf("post message to channel '%s': %w", chanID, err)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return &notification.SentMessage{State: notification.StateSent, StateDetails: "missing users, sent error to channel"}, nil
	}

	// If no users are on-call, we need to abort and let the channel know.
	//
	// This is because we can't update the user group with no members.
	if len(slackUsers) == 0 {
		var buf bytes.Buffer
		err := userGroupErrorEmpty.Execute(&buf, userGroupError{
			GroupID:      ugID,
			ScheduleID:   t.ScheduleID,
			ScheduleName: t.ScheduleName,
			callbackFunc: cfg.CallbackURL,
		})
		if err != nil {
			return nil, fmt.Errorf("execute template: %w", err)
		}
		err = s.withClient(ctx, func(c *slack.Client) error {
			_, _, err := c.PostMessageContext(ctx, chanID, slack.MsgOptionText(buf.String(), false))
			if err != nil {
				return fmt.Errorf("post message to channel '%s': %w", chanID, err)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return &notification.SentMessage{State: notification.StateSent, StateDetails: "empty user-group, sent error to channel"}, nil
	}

	err = s.withClient(ctx, func(c *slack.Client) error {
		_, err := c.UpdateUserGroupMembersContext(ctx, ugID, strings.Join(slackUsers, ","))
		if err != nil {
			return fmt.Errorf("update user group '%s': %w", ugID, err)
		}

		return nil
	})

	// If there was an error, we need to abort and let the channel know.
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
			if err != nil {
				return fmt.Errorf("post message to channel '%s': %w", chanID, err)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return &notification.SentMessage{State: notification.StateSent, StateDetails: "failed to update user-group, sent error to channel and log"}, nil
	}

	return &notification.SentMessage{State: notification.StateDelivered}, nil
}
