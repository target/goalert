package slack

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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

	chanTht *throttle
	listTht *throttle

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

		chanTht: newThrottle(time.Minute / 50),
		listTht: newThrottle(time.Minute / 50),

		listCache: newTTLCache(250, time.Minute),
		chanCache: newTTLCache(1000, 15*time.Minute),
	}, nil
}

func (cs *ChannelSender) client(ctx context.Context) *slack.Client {
	opts := []slack.Option{
		slack.OptionHTTPClient(http.DefaultClient),
	}

	cfg := config.FromContext(ctx)
	if cs.cfg.BaseURL != "" {
		opts = append(opts, slack.OptionAPIURL(cs.cfg.BaseURL))
	}

	return slack.New(cfg.Slack.AccessToken, opts...)
}

// Channel contains information about a Slack channel.
type Channel struct {
	ID     string
	Name   string
	TeamID string
}

type apiError struct {
	msg    string
	header http.Header
}

func (err apiError) Error() string {
	if err.msg == "missing_scope" {
		acceptedScopes := err.header.Get("X-Accepted-Oauth-Scopes")
		providedScopes := err.header.Get("X-Oauth-Scopes")
		return fmt.Sprintf("missing_scope; need one of %v but got %v", acceptedScopes, providedScopes)
	}
	return err.msg
}

func mapError(ctx context.Context, err error) error {
	var apiError *apiError
	if !errors.As(err, &apiError) {
		return err
	}

	switch apiError.msg {
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

	resp, err := s.client(ctx).GetConversationInfoContext(ctx, channelID, false)
	var rateErr *slack.RateLimitedError
	if errors.As(err, &rateErr) {
		s.chanTht.SetWaitUntil(time.Now().Add(rateErr.RetryAfter))
		return s.loadChannel(ctx, channelID)
	}
	if err != nil {
		return nil, err
	}

	return &Channel{
		ID:     resp.ID,
		Name:   "#" + resp.Name,
		TeamID: teamID,
	}, nil
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
	client := s.client(ctx)

	n := 0
	var channels []Channel
	var cursor string
	for {
		n++
		if n > 10 {
			return nil, errors.New("abort after > 10 pages of Slack channels")
		}

		err := s.listTht.Wait(ctx)
		if err != nil {
			return nil, err
		}

		respChan, nextCursor, err := client.GetConversationsForUserContext(ctx, &slack.GetConversationsForUserParameters{
			ExcludeArchived: true,
			Types:           []string{"private_channel", "public_channel"},
			Limit:           200,
			Cursor:          cursor,
		})

		var throttleErr slack.RateLimitedError
		if errors.As(err, &throttleErr) {
			s.listTht.SetWaitUntil(time.Now().Add(throttleErr.RetryAfter))
			continue
		}
		if err != nil {
			return nil, err
		}

		for _, ch := range respChan {
			channels = append(channels, Channel{
				ID:     ch.ID,
				Name:   "#" + ch.Name,
				TeamID: teamID,
			})
		}

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	return channels, nil
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
				slack.MsgOptionText("Broadcasting to channel due to repeat notification.", true),
			)
			break
		}

		opts = append(opts, slack.MsgOptionText(fmt.Sprintf("Alert: %s\n\n<%s>", slackutilsx.EscapeMessage(t.Summary), cfg.CallbackURL("/alerts/"+strconv.Itoa(t.AlertID))), false))
	case notification.AlertStatus:
		opts = append(opts, slack.MsgOptionTS(t.OriginalStatus.ProviderMessageID.ExternalID))

		var status string
		switch t.NewAlertStatus {
		case alert.StatusActive:
			status = "Acknowledged"
		case alert.StatusTriggered:
			status = "Unacknowledged"
		case alert.StatusClosed:
			status = "Closed"
		}

		text := "Status Update: " + status + "\n" + t.LogEntry
		opts = append(opts, slack.MsgOptionText(text, true))
	case notification.AlertBundle:
		opts = append(opts, slack.MsgOptionText(
			fmt.Sprintf("Service '%s' has %d unacknowledged alerts.\n\n<%s>", slackutilsx.EscapeMessage(t.ServiceName), t.Count, cfg.CallbackURL("/services/"+t.ServiceID+"/alerts")),
			false))
	case notification.ScheduleOnCallUsers:
		opts = append(opts, slack.MsgOptionText(s.onCallNotificationText(ctx, t), false))
	default:
		return nil, errors.Errorf("unsupported message type: %T", t)
	}

	_, msgTS, err := s.client(ctx).PostMessageContext(ctx, msg.Destination().Value, opts...)
	if err != nil {
		return nil, err
	}

	return &notification.SentMessage{
		ExternalID: msgTS,
		State:      notification.StateDelivered,
	}, nil
}

func (s *ChannelSender) lookupTeamIDForToken(ctx context.Context, token string) (string, error) {
	resp, err := s.client(ctx).AuthTestContext(ctx)
	if err != nil {
		return "", err
	}

	return resp.TeamID, nil
}
