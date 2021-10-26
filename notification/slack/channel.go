package slack

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
)

type ChannelSender struct {
	cfg Config

	teamID string
	token  string

	chanCache *ttlCache
	listCache *ttlCache

	listMx sync.Mutex
	chanMx sync.Mutex
	teamMx sync.Mutex
}

var _ notification.Sender = &ChannelSender{}

func NewChannelSender(ctx context.Context, cfg Config) (*ChannelSender, error) {
	return &ChannelSender{
		cfg: cfg,

		listCache: newTTLCache(250, time.Minute),
		chanCache: newTTLCache(1000, 15*time.Minute),
	}, nil
}

// Channel contains information about a Slack channel.
type Channel struct {
	ID     string
	Name   string
	TeamID string
}

func rootMsg(err error) string {
	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		return err.Error()
	}

	return rootMsg(unwrapped)
}

func mapError(ctx context.Context, err error) error {
	switch rootMsg(err) {
	case "channel_not_found":
		return validation.NewFieldError("ChannelID", "Invalid Slack channel ID.")
	case "missing_scope", "invalid_auth", "account_inactive", "token_revoked", "not_authed":
		log.Log(ctx, err)
		return validation.NewFieldError("ChannelID", "Permission Denied.")
	}

	return err
}

// Channel will lookup a single Slack channel for the bot.
func (s *ChannelSender) Channel(ctx context.Context, channelID string) (*Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return nil, err
	}

	s.chanMx.Lock()
	defer s.chanMx.Unlock()
	res, ok := s.chanCache.Get(channelID)
	if !ok {
		ch, err := s.loadChannel(ctx, channelID)
		if err != nil {
			return nil, mapError(ctx, err)
		}
		s.chanCache.Add(channelID, ch)
		return ch, nil
	}
	if err != nil {
		return nil, err
	}

	return res.(*Channel), nil
}

func (s *ChannelSender) TeamID(ctx context.Context) (string, error) {
	cfg := config.FromContext(ctx)

	s.teamMx.Lock()
	defer s.teamMx.Unlock()
	if s.teamID == "" || s.token != cfg.Slack.AccessToken {
		// teamID missing or token changed
		id, err := s.lookupTeamIDForToken(ctx, cfg.Slack.AccessToken)
		if err != nil {
			return "", err
		}

		// update teamID and token after fetching succeeds
		s.teamID = id
		s.token = cfg.Slack.AccessToken
	}

	return s.teamID, nil
}

func (s *ChannelSender) loadChannel(ctx context.Context, channelID string) (*Channel, error) {
	teamID, err := s.TeamID(ctx)
	if err != nil {
		return nil, fmt.Errorf("lookup team ID: %w", err)
	}

	ch := &Channel{TeamID: teamID}
	err = s.withClient(ctx, func(c *slack.Client) error {
		resp, err := c.GetConversationInfoContext(ctx, channelID, false)
		if err != nil {
			return err
		}

		ch.ID = resp.ID
		ch.Name = "#" + resp.Name

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("lookup conversation info: %w", err)
	}

	return ch, nil
}

// ListChannels will return a list of channels visible to the slack bot.
func (s *ChannelSender) ListChannels(ctx context.Context) ([]Channel, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return nil, err
	}

	cfg := config.FromContext(ctx)
	s.listMx.Lock()
	defer s.listMx.Unlock()
	res, ok := s.listCache.Get(cfg.Slack.AccessToken)
	if !ok {
		chs, err := s.loadChannels(ctx)
		if err != nil {
			return nil, mapError(ctx, err)
		}
		ch2 := make([]Channel, len(chs))
		copy(ch2, chs)
		s.listCache.Add(cfg.Slack.AccessToken, ch2)
		return chs, nil
	}
	if err != nil {
		return nil, err
	}

	chs := res.([]Channel)
	cpy := make([]Channel, len(chs))
	copy(cpy, chs)

	return cpy, nil
}

func (s *ChannelSender) loadChannels(ctx context.Context) ([]Channel, error) {
	teamID, err := s.TeamID(ctx)
	if err != nil {
		return nil, fmt.Errorf("lookup team ID: %w", err)
	}

	n := 0
	var channels []Channel
	var cursor string
	for {
		n++
		if n > 10 {
			return nil, errors.New("abort after > 10 pages of Slack channels")
		}

		err = s.withClient(ctx, func(c *slack.Client) error {
			respChan, nextCursor, err := c.GetConversationsForUserContext(ctx, &slack.GetConversationsForUserParameters{
				ExcludeArchived: true,
				Types:           []string{"private_channel", "public_channel"},
				Limit:           200,
				Cursor:          cursor,
			})
			if err != nil {
				return err
			}

			cursor = nextCursor

			for _, ch := range respChan {
				channels = append(channels, Channel{
					ID:     ch.ID,
					Name:   "#" + ch.Name,
					TeamID: teamID,
				})
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("list channels: %w", err)
		}

		if cursor == "" {
			break
		}
	}

	return channels, nil
}

func alertLink(ctx context.Context, id int, summary string) string {
	cfg := config.FromContext(ctx)
	path := fmt.Sprintf("/alerts/%d", id)
	return fmt.Sprintf("<%s|Alert #%d: %s>", cfg.CallbackURL(path), id, slackutilsx.EscapeMessage(summary))
}

func (s *ChannelSender) Send(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {

	cfg := config.FromContext(ctx)

	// Note: We don't use cfg.ApplicationName() here since that is configured in the Slack app as the bot name.

	var opts []slack.MsgOption
	switch t := msg.(type) {
	case notification.Alert:
		if t.OriginalStatus != nil {
			// Reply in thread if we already sent a message for this alert.
			opts = append(opts,
				slack.MsgOptionTS(t.OriginalStatus.ProviderMessageID.ExternalID),
				slack.MsgOptionBroadcast(),
				slack.MsgOptionText(alertLink(ctx, t.AlertID, t.Summary), false),
			)
			break
		}

		details := "> " + strings.ReplaceAll(slackutilsx.EscapeMessage(t.Details), "\n", "\n> ")
		opts = append(opts,
			slack.MsgOptionAttachments(
				slack.Attachment{
					Color: "danger",
					Blocks: slack.Blocks{
						BlockSet: []slack.Block{
							slack.NewTextBlockObject("mrkdwn", alertLink(ctx, t.AlertID, t.Summary), false, false),
							slack.NewTextBlockObject("mrkdwn", details, false, false),
							slack.NewContextBlock("", slack.NewTextBlockObject("plain_text", "Unacknowledged", false, false)),
						},
					},
				},
			),
		)
	case notification.AlertStatus:
		var color string
		details := "> " + strings.ReplaceAll(slackutilsx.EscapeMessage(t.Details), "\n", "\n> ")
		switch t.NewAlertStatus {
		case alert.StatusActive:
			color = "warning"
		case alert.StatusTriggered:
			color = "danger"
		case alert.StatusClosed:
			color = "good"
			details = ""
		}

		opts = append(opts,
			slack.MsgOptionUpdate(t.OriginalStatus.ProviderMessageID.ExternalID),
			slack.MsgOptionAttachments(
				slack.Attachment{
					Color: color,
					Blocks: slack.Blocks{
						BlockSet: []slack.Block{
							slack.NewTextBlockObject("mrkdwn", alertLink(ctx, t.AlertID, t.Summary), false, false),
							slack.NewTextBlockObject("mrkdwn", details, false, false),
							slack.NewContextBlock("", slack.NewTextBlockObject("plain_text", slackutilsx.EscapeMessage(t.LogEntry), false, false)),
						},
					},
				},
			),
		)
	case notification.AlertBundle:
		opts = append(opts, slack.MsgOptionText(
			fmt.Sprintf("Service '%s' has %d unacknowledged alerts.\n\n<%s>", slackutilsx.EscapeMessage(t.ServiceName), t.Count, cfg.CallbackURL("/services/"+t.ServiceID+"/alerts")),
			false))
	case notification.ScheduleOnCallUsers:
		opts = append(opts, slack.MsgOptionText(s.onCallNotificationText(ctx, t), false))
	default:
		return nil, errors.Errorf("unsupported message type: %T", t)
	}

	var msgTS string
	err := s.withClient(ctx, func(c *slack.Client) error {
		_, _msgTS, err := c.PostMessageContext(ctx, msg.Destination().Value, opts...)
		if err != nil {
			return err
		}
		msgTS = _msgTS
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &notification.SentMessage{
		ExternalID: msgTS,
		State:      notification.StateDelivered,
	}, nil
}

func (s *ChannelSender) lookupTeamIDForToken(ctx context.Context, token string) (string, error) {
	var teamID string

	err := s.withClient(ctx, func(c *slack.Client) error {
		info, err := c.AuthTestContext(ctx)
		if err != nil {
			return err
		}

		teamID = info.TeamID

		return nil
	})

	if err != nil {
		return "", err
	}

	return teamID, nil
}
