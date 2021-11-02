package slack

import (
	"encoding/json"
	"net/http"

	"github.com/target/goalert/notification"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/validation"
)

func (s *ChannelSender) ServeMessageAction(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var payload struct {
		Type string
		User struct {
			ID     string `json:"id"`
			TeamID string `json:"team_id"`
		}
		Actions []struct {
			ActionID string `json:"action_id"`
			BlockID  string `json:"block_id"`
			Value    string `json:"value"`
		}
	}
	err := json.Unmarshal([]byte(req.FormValue("payload")), &payload)
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
	default:
		errutil.HTTPError(ctx, w, validation.NewFieldErrorf("action_id", "unknown action ID '%s'", act.ActionID))
		return
	}

	err = s.recv.ReceiveSubject(ctx, "slack:"+payload.User.TeamID, payload.User.ID, act.Value, res)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
}
