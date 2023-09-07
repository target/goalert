package slack

import (
	"bytes"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/auth/authlink"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
)

func validateRequestSignature(now time.Time, req *http.Request) error {
	cfg := config.FromContext(req.Context())

	if req.Form != nil {
		return errors.New("request already parsed, can't validate signature")
	}

	// copy body data
	var buf bytes.Buffer
	if req.Body != nil {
		orig := req.Body
		req.Body = io.NopCloser(io.TeeReader(req.Body, &buf))
		err := req.ParseForm()
		orig.Close()
		if err != nil {
			return err
		}
	}

	// read ts
	tsStr := req.Header.Get("X-Slack-Request-Timestamp")
	unixSec, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return permission.Unauthorized()
	}
	ts := time.Unix(unixSec, 0)
	if now.Sub(ts).Abs() > 5*time.Minute {
		return permission.Unauthorized()
	}

	properSig := Signature(cfg.Slack.SigningSecret, ts, buf.Bytes())
	if !hmac.Equal([]byte(req.Header.Get("X-Slack-Signature")), []byte(properSig)) {
		return permission.Unauthorized()
	}

	return nil
}

func (s *ChannelSender) ServeMessageAction(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	cfg := config.FromContext(ctx)

	if !cfg.Slack.InteractiveMessages {
		http.Error(w, "not enabled", http.StatusNotFound)
		return
	}

	err := validateRequestSignature(time.Now(), req)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	var payload struct {
		Type        string
		ResponseURL string `json:"response_url"`
		Team        struct {
			ID     string
			Domain string
		}
		Channel struct {
			ID string
		}
		User struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Name     string
		}
		Actions []struct {
			ActionID string `json:"action_id"`
			BlockID  string `json:"block_id"`
			Value    string `json:"value"`
		}
	}
	err = json.Unmarshal([]byte(req.FormValue("payload")), &payload)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	if len(payload.Actions) != 1 {
		errutil.HTTPError(ctx, w, validation.NewFieldError("payload", "invalid payload"))
		return
	}

	act := payload.Actions[0]
	if act.BlockID != alertResponseBlockID {
		errutil.HTTPError(ctx, w, validation.NewFieldErrorf("block_id", "unknown block ID '%s'", act.BlockID))
		return
	}

	var res notification.Result
	switch act.ActionID {
	case alertAckActionID:
		res = notification.ResultAcknowledge
	case alertCloseActionID:
		res = notification.ResultResolve
	case linkActActionID:
		err = s.withClient(ctx, func(c *slack.Client) error {
			// remove ephemeral 'Link Account' button
			_, err = c.PostEphemeralContext(ctx, payload.Channel.ID, payload.User.ID,
				slack.MsgOptionText("", false), slack.MsgOptionReplaceOriginal(payload.ResponseURL),
				slack.MsgOptionDeleteOriginal(payload.ResponseURL))
			if err != nil {
				return fmt.Errorf("delete ephemeral message: %w", err)
			}
			return nil
		})
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		return
	default:
		errutil.HTTPError(ctx, w, validation.NewFieldErrorf("action_id", "unknown action ID '%s'", act.ActionID))
		return
	}

	var e *notification.UnknownSubjectError
	err = s.recv.ReceiveSubject(ctx, "slack:"+payload.Team.ID, payload.User.ID, act.Value, res)

	if errors.As(err, &e) {
		var linkURL string
		switch {
		case payload.User.Name == "", payload.User.Username == "", payload.Team.ID == "", payload.Team.Domain == "":
			// missing data, don't allow linking
			log.Log(ctx, errors.New("slack payload missing required data"))
		default:
			linkURL, err = s.recv.AuthLinkURL(ctx, "slack:"+payload.Team.ID, payload.User.ID, authlink.Metadata{
				UserDetails: fmt.Sprintf("Slack user %s (@%s) from %s.slack.com", payload.User.Name, payload.User.Username, payload.Team.Domain),
				AlertID:     e.AlertID,
				AlertAction: res.String(),
			})
			if err != nil {
				log.Log(ctx, err)
			}
		}

		err = s.withClient(ctx, func(c *slack.Client) error {
			var msg string
			if linkURL == "" {
				msg = "Your Slack account isn't currently linked to GoAlert, please try again later."
			} else {
				msg = "Please link your Slack account with GoAlert."
			}
			blocks := []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("plain_text", msg, false, false),
					nil, nil,
				),
			}

			if linkURL != "" {
				btn := slack.NewButtonBlockElement(linkActActionID, linkURL,
					slack.NewTextBlockObject("plain_text", "Link Account", false, false))
				btn.URL = linkURL
				blocks = append(blocks, slack.NewActionBlock(alertResponseBlockID, btn))
			}

			_, err = c.PostEphemeralContext(ctx, payload.Channel.ID, payload.User.ID,
				slack.MsgOptionResponseURL(payload.ResponseURL, "ephemeral"),
				slack.MsgOptionBlocks(blocks...),
			)
			if err != nil {
				return err
			}
			return nil
		})
		return
	}
	if alert.IsAlreadyAcknowledged(err) || alert.IsAlreadyClosed(err) {
		// ignore errors from duplicate requests
		return
	}
	if errutil.HTTPError(ctx, w, err) {
		return
	}
}
