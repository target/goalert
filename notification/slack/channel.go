package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
	"golang.org/x/net/context/ctxhttp"
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
	cfg := config.FromContext(ctx)

	teamID, err := s.TeamID(ctx)
	if err != nil {
		return nil, fmt.Errorf("lookup team ID: %w", err)
	}

	v := make(url.Values)
	// Parameters and URL documented here:
	// https://api.slack.com/methods/conversations.info
	v.Set("token", cfg.Slack.AccessToken)
	v.Set("channel", channelID)

	infoURL := s.cfg.url("/api/conversations.info")

	var resData struct {
		OK      bool
		Error   string
		Channel struct {
			ID   string
			Name string
		}
	}

	err = s.chanTht.Wait(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := ctxhttp.PostForm(ctx, http.DefaultClient, infoURL, v)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 429 {
		// respect Retry-After (seconds) if possible
		sec, err := strconv.Atoi(resp.Header.Get("Retry-After"))
		if err == nil {
			s.chanTht.SetWaitUntil(time.Now().Add(time.Second * time.Duration(sec)))
			// try again
			return s.loadChannel(ctx, channelID)
		}
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, errors.New("non-200 response from Slack: " + resp.Status)
	}
	err = json.NewDecoder(resp.Body).Decode(&resData)
	resp.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "parse JSON")
	}

	if !resData.OK {
		return nil, fmt.Errorf("lookup Slack channel: %w", &apiError{msg: resData.Error, header: resp.Header})
	}

	return &Channel{
		ID:     resData.Channel.ID,
		Name:   "#" + resData.Channel.Name,
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
	cfg := config.FromContext(ctx)
	v := make(url.Values)
	// Parameters and URL documented here:
	// https://api.slack.com/methods/users.conversations
	v.Set("token", cfg.Slack.AccessToken)
	v.Set("exclude_archived", "true")

	// Using `Set` instead of `Add` here. Slack expects a comma-delimited list instead of
	// an array-encoded parameter.
	v.Set("types", "private_channel,public_channel")
	v.Set("limit", "200")
	listURL := s.cfg.url("/api/users.conversations")

	n := 0
	var channels []Channel
	for {
		n++
		if n > 10 {
			return nil, errors.New("abort after > 10 pages of Slack channels")
		}

		err := s.listTht.Wait(ctx)
		if err != nil {
			return nil, err
		}
		resp, err := ctxhttp.PostForm(ctx, http.DefaultClient, listURL, v)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 429 {
			resp.Body.Close()
			// respect Retry-After (seconds) if possible
			sec, err := strconv.Atoi(resp.Header.Get("Retry-After"))
			if err == nil {
				s.listTht.SetWaitUntil(time.Now().Add(time.Second * time.Duration(sec)))
				// no need to start over, re-fetch current page
				continue
			}
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, errors.New("non-200 response from Slack: " + resp.Status)
		}

		var resData struct {
			OK       bool
			Error    string
			Channels []Channel
			Meta     struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}

		err = json.NewDecoder(resp.Body).Decode(&resData)
		resp.Body.Close()
		if err != nil {
			return nil, errors.Wrap(err, "parse JSON")
		}

		if !resData.OK {
			return nil, fmt.Errorf("list Slack channels: %w", &apiError{msg: resData.Error, header: resp.Header})
		}

		channels = append(channels, resData.Channels...)

		if resData.Meta.NextCursor == "" {
			break
		}

		v.Set("cursor", resData.Meta.NextCursor)
	}

	for i := range channels {
		channels[i].Name = "#" + channels[i].Name
	}

	return channels, nil
}

func (s *ChannelSender) Send(ctx context.Context, msg notification.Message) (string, *notification.Status, error) {
	cfg := config.FromContext(ctx)

	vals := make(url.Values)
	// Parameters & URL documented here:
	// https://api.slack.com/methods/chat.postMessage
	vals.Set("channel", msg.Destination().Value)
	switch t := msg.(type) {
	case notification.Alert:
		if t.OriginalStatus != nil {
			// Reply in thread if we already sent a message for this alert.
			vals.Set("thread_ts", t.OriginalStatus.ProviderMessageID.ExternalID)
			vals.Set("text", "Escalated.")
			vals.Set("reply_broadcast", "true")
			break
		}

		vals.Set("text", fmt.Sprintf("Alert: %s\n\n<%s>", t.Summary, cfg.CallbackURL("/alerts/"+strconv.Itoa(t.AlertID))))
	case notification.AlertBundle:
		vals.Set("text", fmt.Sprintf("Service '%s' has %d unacknowledged alerts.\n\n<%s>", t.ServiceName, t.Count, cfg.CallbackURL("/services/"+t.ServiceID+"/alerts")))
	case notification.ScheduleOnCallUsers:
		var userStr string
		if len(t.Users) == 0 {
			userStr = "No users"
		} else {
			teamID, err := s.TeamID(ctx)
			if err != nil {
				log.Log(ctx, fmt.Errorf("lookup team ID: %w", err))
			}

			m := make(map[string]string, len(t.Users))
			if teamID != "" {
				userIDs := make([]string, len(t.Users))
				for i, u := range t.Users {
					userIDs[i] = u.ID
				}
				err = s.cfg.UserStore.AuthSubjectsFunc(ctx, "slack:"+teamID, func(sub user.AuthSubject) error {
					m[sub.UserID] = sub.SubjectID
					return nil
				}, userIDs...)
				if err != nil {
					log.Log(ctx, fmt.Errorf("lookup auth subjects for slack: %w", err))
				}
			}

			var userLinks []string
			for _, u := range t.Users {
				subjectID := m[u.ID]
				if subjectID == "" {
					// fallback to a link to the GoAlert user
					userLinks = append(userLinks, fmt.Sprintf("<%s|%s>", u.URL, u.Name))
					continue
				}

				userLinks = append(userLinks, fmt.Sprintf("<@%s>", subjectID))
			}
			userStr = "Users: " + strings.Join(userLinks, ", ")
		}

		vals.Set("text", fmt.Sprintf("%s are on-call for Schedule: %s", userStr, fmt.Sprintf("<%s|%s>", t.ScheduleURL, t.ScheduleName)))
	default:
		return "", nil, errors.Errorf("unsupported message type: %T", t)
	}
	vals.Set("token", cfg.Slack.AccessToken)

	resp, err := http.PostForm(s.cfg.url("/api/chat.postMessage"), vals)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", nil, errors.Errorf("non-200 response: %s", resp.Status)
	}

	var resData struct {
		OK    bool
		Error string
		TS    string
	}
	err = json.NewDecoder(resp.Body).Decode(&resData)
	if err != nil {
		return "", nil, errors.Wrap(err, "decode response")
	}
	if !resData.OK {
		return "", nil, errors.Errorf("Slack error: %s", resData.Error)
	}

	return resData.TS, &notification.Status{State: notification.StateDelivered}, nil
}

func (s *ChannelSender) lookupTeamIDForToken(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", s.cfg.url("/api/auth.test"), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body struct {
		TeamID string `json:"team_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", err
	}
	return body.TeamID, nil
}
