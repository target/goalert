package slack

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
)

// Handler responds to API requests from Slack
type Handler struct {
	c Config
}

// Payload represents the relevant payload information sent from Slack
type Payload struct {
	ResponseURL string `json:"response_url"`
	Actions     []Action
	Channel     slack.Channel
}

// Action represents the information given from an action event within Slack
// i.e. clicking to acknowledge an alert from slack
type Action struct {
	ActionID string `json:"action_id"`
	ActionTS string `json:"action_ts"`
	BlockID  string `json:"block_id"`
	Value    string
}

// NewHandler creates a new Handler, registering API routes using chi.
func NewHandler(c Config) *Handler {
	return &Handler{c: c}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// validRequest is used to validate a request from Slack.
// If the request is validated true is returned, false otherwise.
// https://api.slack.com/authentication/verifying-requests-from-slack
func validRequest(w http.ResponseWriter, req *http.Request) bool {
	if req.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return false
	}

	ts := req.Header.Get("X-Slack-Request-Timestamp")

	// ignore request if more than 5 minutes from local time
	_ts, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return false
	}
	if abs(time.Now().Unix()-_ts) > 60*5 {
		return false
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return false
	}
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	secret := config.FromContext(req.Context()).Slack.SigningSecret
	h := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(h, "v0:%s:%s", ts, body)
	calculatedSignature := "v0=" + hex.EncodeToString(h.Sum(nil))
	signature := []byte(req.Header.Get("X-Slack-Signature"))

	return hmac.Equal(signature, []byte(calculatedSignature))
}

// ServeActionCallback processes POST requests from Slack. A callback ID is provided
// to determine which action to take.
func (h *Handler) ServeActionCallback(w http.ResponseWriter, req *http.Request) {
	if !validRequest(w, req) {
		fmt.Println("request invalid")
		return
	}

	payload := req.FormValue("payload")
	p := Payload{}
	json.Unmarshal([]byte(payload), &p)

	writeHTTPErr := func(err error) {
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}

	process := func(ctx context.Context) {
		for _, a := range p.Actions {
			v, err := strconv.Atoi(a.Value)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
				return
			}

			switch a.ActionID {
			case "ack":
				err := h.c.AlertStore.UpdateStatus(ctx, v, alert.StatusActive)
				writeHTTPErr(err)
			case "esc":
				err := h.c.AlertStore.Escalate(ctx, v)
				writeHTTPErr(err)
			case "close":
				err := h.c.AlertStore.UpdateStatus(ctx, v, alert.StatusClosed)
				writeHTTPErr(err)
			case "open":
			}
		}
	}

	permission.SudoContext(req.Context(), process)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) _ServeActionCallback(w http.ResponseWriter, req *http.Request) {
	writeHTTPErr := func(err error) {
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	if !validRequest(w, req) {
		fmt.Println("request invalid")
		return
	}

	payload := req.FormValue("payload")
	p := Payload{}
	json.Unmarshal([]byte(payload), &p)

	process := func(ctx context.Context) {
		cfg := config.FromContext(ctx)
		var api = slack.New(cfg.Slack.AccessToken)
		for _, a := range p.Actions {
			v, err := strconv.Atoi(a.Value)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
				return
			}

			switch a.ActionID {
			case "ack":
				err := h.c.AlertStore.UpdateStatus(ctx, v, alert.StatusActive)
				var1, var2, var3 := api.PostMessage(p.Channel.ID, slack.MsgOptionText("Yes, hello.", false))
				fmt.Println(var1, var2, var3)
				writeHTTPErr(err)

			case "esc":
				err := h.c.AlertStore.Escalate(ctx, v)
				writeHTTPErr(err)
			case "close":
				err := h.c.AlertStore.UpdateStatus(ctx, v, alert.StatusClosed)
				writeHTTPErr(err)
			case "open":
			}
		}
	}

	permission.SudoContext(req.Context(), process)
}
