package slack

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackutilsx"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
)

type ChannelSender struct {
	cfg Config

	teamID string
	token  string

	chanCache *ttlCache[string, *Channel]
	listCache *ttlCache[string, []Channel]
	ugCache   *ttlCache[string, []slack.UserGroup]

	teamInfoCache *ttlCache[string, *slack.TeamInfo]
	userInfoCache *ttlCache[string, *slack.User]
	ugInfoCache   *ttlCache[string, UserGroup]

	listMx sync.Mutex
	chanMx sync.Mutex
	teamMx sync.Mutex

	teamInfoMx sync.Mutex

	recv notification.Receiver
}

const (
	colorClosed  = "#218626"
	colorUnacked = "#862421"
	colorAcked   = "#867321"
)

var (
	_ notification.Sender         = &ChannelSender{}
	_ notification.ReceiverSetter = &ChannelSender{}
)

func NewChannelSender(ctx context.Context, cfg Config) (*ChannelSender, error) {
	return &ChannelSender{
		cfg: cfg,

		listCache: newTTLCache[string, []Channel](250, time.Minute),
		chanCache: newTTLCache[string, *Channel](1000, 15*time.Minute),
		ugCache:   newTTLCache[string, []slack.UserGroup](1000, time.Minute),

		teamInfoCache: newTTLCache[string, *slack.TeamInfo](1, 15*time.Minute),
		userInfoCache: newTTLCache[string, *slack.User](1000, 15*time.Minute),
		ugInfoCache:   newTTLCache[string, UserGroup](1000, 15*time.Minute),
	}, nil
}

func (s *ChannelSender) SetReceiver(r notification.Receiver) {
	s.recv = r
}

// Channel contains information about a Slack channel.
type Channel struct {
	ID     string
	Name   string
	TeamID string

	IsArchived bool
}

func (c Channel) AsField() nfydest.FieldValue {
	return nfydest.FieldValue{
		Value: c.ID,
		Label: c.Name,
	}
}

// User contains information about a Slack user.
type User struct {
	ID     string
	Name   string
	TeamID string
}

// Team contains information about a Slack team.
type Team struct {
	ID      string
	Domain  string
	Name    string
	IconURL string
}

func (t Team) ChannelLink(id string) string {
	var u url.URL

	u.Host = t.Domain + ".slack.com"
	u.Scheme = "https"
	u.Path = "/archives/" + url.PathEscape(id)

	return u.String()
}

func (t Team) UserLink(id string) string {
	var u url.URL

	u.Host = t.Domain + ".slack.com"
	u.Scheme = "https"
	u.Path = "/team/" + url.PathEscape(id)

	return u.String()
}

func rootMsg(err error) string {
	if err == nil {
		return ""
	}
	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		return err.Error()
	}

	return rootMsg(unwrapped)
}

func mapError(ctx context.Context, err error) error {
	switch rootMsg(err) {
	case "channel_not_found":
		return validation.NewFieldError("ChannelID", "Channel does not exist, is archived, or is private (invite goalert bot).")
	case "missing_scope", "invalid_auth", "account_inactive", "token_revoked", "not_authed":
		log.Log(ctx, err)
		return validation.NewFieldError("ChannelID", "Permission Denied.")
	}

	return err
}

func (s *ChannelSender) ValidateChannel(ctx context.Context, id string) error {
	err := permission.LimitCheckAny(ctx, permission.User, permission.System)
	if err != nil {
		return err
	}

	if id == "" {
		return validation.NewGenericError("Channel is required.")
	}

	s.chanMx.Lock()
	defer s.chanMx.Unlock()
	res, ok := s.chanCache.Get(id)
	if !ok {
		res, err = s.loadChannel(ctx, id)
		if err != nil {
			if rootMsg(err) == "channel_not_found" {
				return validation.NewGenericError("Channel does not exist, is archived, or is private (invite goalert bot).")
			}

			return err
		}
		s.chanCache.Add(id, res)
	}

	if res.IsArchived {
		return validation.NewGenericError("Channel is archived.")
	}

	return nil
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

	return res, nil
}

func (s *ChannelSender) Team(ctx context.Context, id string) (t *Team, err error) {
	s.teamInfoMx.Lock()
	defer s.teamInfoMx.Unlock()

	info, ok := s.teamInfoCache.Get(id)
	if ok {
		url, _ := info.Icon["image_44"].(string)
		return &Team{
			ID:      info.ID,
			Name:    info.Name,
			IconURL: url,
			Domain:  info.Domain,
		}, nil
	}

	err = s.withClient(ctx, func(c *slack.Client) error {
		info, err := c.GetTeamInfoContext(ctx)
		if err != nil {
			return err
		}

		url, _ := info.Icon["image_44"].(string)
		t = &Team{
			ID:      info.ID,
			Name:    info.Name,
			IconURL: url,
			Domain:  info.Domain,
		}

		s.teamInfoCache.Add(id, info)
		return nil
	})

	return t, err
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
		resp, err := c.GetConversationInfoContext(ctx,
			&slack.GetConversationInfoInput{
				ChannelID: channelID,
			})
		if err != nil {
			return err
		}

		ch.ID = resp.ID
		ch.Name = "#" + resp.Name
		ch.IsArchived = resp.IsArchived

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

	cpy := make([]Channel, len(res))
	copy(cpy, res)

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

			for _, rCh := range respChan {
				ch := Channel{
					ID:     rCh.ID,
					Name:   "#" + rCh.Name,
					TeamID: teamID,
				}
				channels = append(channels, ch)
				s.chanCache.Add(ch.ID, &ch)
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

const (
	alertResponseBlockID = "block_alert_response"
	alertCloseActionID   = "action_alert_close"
	alertAckActionID     = "action_alert_ack"
	linkActActionID      = "action_link_account"
)

// alertMsgOption will return the slack.MsgOption for an alert-type message (e.g., notification or status update).
func alertMsgOption(ctx context.Context, callbackID string, id int, summary, logEntry string, state notification.AlertState) slack.MsgOption {
	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", alertLink(ctx, id, summary), false, false), nil, nil),
	}

	var color string
	var actions []slack.Block
	switch state {
	case notification.AlertStateAcknowledged:
		color = colorAcked
		actions = []slack.Block{
			slack.NewDividerBlock(),
			slack.NewActionBlock(alertResponseBlockID,
				slack.NewButtonBlockElement(alertCloseActionID, callbackID, slack.NewTextBlockObject("plain_text", "Close", false, false)),
			),
		}
	case notification.AlertStateUnacknowledged:
		color = colorUnacked
		actions = []slack.Block{
			slack.NewDividerBlock(),
			slack.NewActionBlock(alertResponseBlockID,
				slack.NewButtonBlockElement(alertAckActionID, callbackID, slack.NewTextBlockObject("plain_text", "Acknowledge", false, false)),
				slack.NewButtonBlockElement(alertCloseActionID, callbackID, slack.NewTextBlockObject("plain_text", "Close", false, false)),
			),
		}
	case notification.AlertStateClosed:
		color = colorClosed
	}

	blocks = append(blocks,
		slack.NewContextBlock("", slack.NewTextBlockObject("plain_text", logEntry, false, false)),
	)
	cfg := config.FromContext(ctx)
	if len(actions) > 0 && cfg.Slack.InteractiveMessages {
		blocks = append(blocks, actions...)
	}

	return slack.MsgOptionAttachments(
		slack.Attachment{
			Color:    color,
			Fallback: fmt.Sprintf("Alert #%d: %s", id, slackutilsx.EscapeMessage(summary)),
			Blocks:   slack.Blocks{BlockSet: blocks},
		},
	)
}

func chanTS(origChannelID, externalID string) (channelID, ts string) {
	ts = externalID
	if strings.Contains(ts, ":") {
		// DMs have a channel ID and timestamp separated by a colon,
		// so we need to split them out. Trying to update a message
		// with a user ID will fail.
		channelID, ts, _ = strings.Cut(ts, ":")
	} else {
		channelID = origChannelID
	}

	return channelID, ts
}

func (s *ChannelSender) Send(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	cfg := config.FromContext(ctx)

	// Note: We don't use cfg.ApplicationName() here since that is configured in the Slack app as the bot name.

	var opts []slack.MsgOption
	var isUpdate bool
	channelID := msg.DestArg(FieldSlackChannelID)
	if msg.DestType() == DestTypeSlackDirectMessage {
		// DMs are sent to the user ID, not the channel ID.
		channelID = msg.DestArg(FieldSlackUserID)
	}

	switch t := msg.(type) {
	case notification.Test:
		opts = append(opts, slack.MsgOptionText("This is a test message.", false))
	case notification.Verification:
		opts = append(opts, slack.MsgOptionText(fmt.Sprintf("Your verification code is: %s", t.Code), false))
	case notification.Alert:
		if t.OriginalStatus != nil {
			var ts string
			channelID, ts = chanTS(channelID, t.OriginalStatus.ProviderMessageID.ExternalID)

			// Reply in thread if we already sent a message for this alert.
			opts = append(opts,
				slack.MsgOptionTS(ts),
				slack.MsgOptionBroadcast(),
				slack.MsgOptionText(alertLink(ctx, t.AlertID, t.Summary), false),
			)
			break
		}

		opts = append(opts, alertMsgOption(ctx, t.MsgID(), t.AlertID, t.Summary, "Unacknowledged", notification.AlertStateUnacknowledged))
	case notification.AlertStatus:
		isUpdate = true
		var ts string
		channelID, ts = chanTS(channelID, t.OriginalStatus.ProviderMessageID.ExternalID)
		opts = append(opts,
			slack.MsgOptionUpdate(ts),
			alertMsgOption(ctx, t.OriginalStatus.ID, t.AlertID, t.Summary, t.LogEntry, t.NewAlertState),
		)
	case notification.AlertBundle:
		opts = append(opts, slack.MsgOptionText(
			fmt.Sprintf("Service '%s' has %d unacknowledged alerts.\n\n<%s>", slackutilsx.EscapeMessage(t.ServiceName), t.Count, cfg.CallbackURL("/services/"+t.ServiceID+"/alerts")),
			false))
	case notification.SignalMessage:
		opts = append(opts, slack.MsgOptionText(t.Param("message"), false))
	case notification.ScheduleOnCallUsers:
		opts = append(opts, slack.MsgOptionText(s.onCallNotificationText(ctx, t), false))
	default:
		return nil, errors.Errorf("unsupported message type: %T", t)
	}

	var externalID string
	err := s.withClient(ctx, func(c *slack.Client) error {
		msgChan, msgTS, err := c.PostMessageContext(ctx, channelID, opts...)
		if err != nil {
			return err
		}
		if msgChan != channelID {
			// DMs have a generated channel ID that we need to store
			// along with the timestamp that does not match the original
			// in order to update the message.
			externalID = fmt.Sprintf("%s:%s", msgChan, msgTS)
		} else {
			// For other channels, we can just store the timestamp,
			// to preserve compatibility with older versions of GoAlert.
			externalID = msgTS
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if isUpdate {
		externalID = ""
	}

	return &notification.SentMessage{
		ExternalID: externalID,
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
