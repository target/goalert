package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
)

func validateRequestSignature(req *http.Request) error {
	cfg := config.FromContext(req.Context())

	var newBody struct {
		io.Reader
		io.Closer
	}

	h := hmac.New(sha256.New, []byte(cfg.Slack.SigningSecret))
	tsStr := req.Header.Get("X-Slack-Request-Timestamp")
	io.WriteString(h, "v0:"+tsStr+":")
	newBody.Reader = io.TeeReader(req.Body, h)
	newBody.Closer = req.Body
	req.Body = newBody

	err := req.ParseForm()
	if err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}
	diff := time.Since(time.Unix(ts, 0))
	if diff < 0 {
		diff = -diff
	}
	if diff > 5*time.Minute {
		return fmt.Errorf("timestamp too old: %s", diff)
	}

	sig := "v0=" + hex.EncodeToString(h.Sum(nil))
	if hmac.Equal([]byte(req.Header.Get("X-Slack-Signature")), []byte(sig)) {
		return fmt.Errorf("invalid signature")
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

	err := validateRequestSignature(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		s.withClient(ctx, func(c *slack.Client) error {
			// remove ephemeral 'Link Account' button
			_, err = c.PostEphemeralContext(ctx, payload.Channel.ID, payload.User.ID,
				slack.MsgOptionText("", false), slack.MsgOptionReplaceOriginal(payload.ResponseURL),
				slack.MsgOptionDeleteOriginal(payload.ResponseURL))
			if err != nil {
				return err
			}
			return nil
		})
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
				msg = "Please link your Slack account with GoAlert then try again."
			}

			btn := slack.NewButtonBlockElement(linkActActionID, linkURL,
				slack.NewTextBlockObject("plain_text", "Link Account", false, false))
			btn.URL = linkURL

			_, err = c.PostEphemeralContext(ctx, payload.Channel.ID, payload.User.ID,
				slack.MsgOptionResponseURL(payload.ResponseURL, "ephemeral"),
				slack.MsgOptionBlocks(
					slack.NewSectionBlock(
						slack.NewTextBlockObject("plain_text", msg, false, false),
						nil, nil,
					),
					slack.NewActionBlock(alertResponseBlockID, btn),
				),
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
