package twilio

import (
	"context"
	"database/sql"
	"encoding/base64"
	stderrors "errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
)

// CallType indicates a supported Twilio voice call type.
type CallType string

// KeyPressed specifies a key pressed from the voice menu options.
type KeyPressed string

// Voice implements a notification.Sender for Twilio voice calls.
type Voice struct {
	c *Config
	r notification.Receiver
}

const (
	// Supported call types.
	CallTypeAlert       = CallType("alert")
	CallTypeAlertStatus = CallType("alert-status")
	CallTypeTest        = CallType("test")
	CallTypeVerify      = CallType("verify")
	CallTypeStop        = CallType("stop")

	// Possible keys pressed from the Menu mapped to their actions.
	digitAck      = "4"
	digitClose    = "6"
	digitStop     = "1"
	digitGoBack   = "1"
	digitRepeat   = "*"
	digitConfirm  = "3"
	digitOldAck   = "8"
	digitOldClose = "9"
	digitEscalate = "5"
	sayRepeat     = "star"
)

var (
	// We use url encoding with no padding to try and eliminate
	// encoding problems with buggy apps.
	b64enc = base64.URLEncoding.WithPadding(base64.NoPadding)

	errVoiceTimeout                             = errors.New("process voice action: timeout")
	pRx                                         = regexp.MustCompile(`\((.*?)\)`)
	_               notification.ReceiverSetter = &Voice{}
	_               notification.Sender         = &Voice{}
	_               notification.StatusChecker  = &Voice{}
	_               notification.FriendlyValuer = &Voice{}
	rmParen                                     = regexp.MustCompile(`\s*\(.*?\)`)
)

func voiceErrorMessage(ctx context.Context, err error) (string, error) {
	var e alert.LogEntryFetcher
	if errors.As(err, &e) {
		// we pass a 'sudo' context to give permission
		var msg string
		permission.SudoContext(ctx, func(sCtx context.Context) {
			entry, err := e.LogEntry(sCtx)
			if err != nil {
				log.Log(sCtx, errors.Wrap(err, "fetch log entry"))
			} else {
				// Stripping off anything in between parenthesis
				msg = "Already " + pRx.ReplaceAllString(entry.String(ctx), "")
			}
		})
		if msg != "" {
			return msg, nil
		}
	}
	// In case we don't get a log entry, respond with generic messages.
	if alert.IsAlreadyClosed(err) {
		return "Alert is already closed.", nil
	}
	if alert.IsAlreadyAcknowledged(err) {
		return "Alert is already acknowledged.", nil
	}
	if validation.IsClientError(err) {
		return "Error: " + stderrors.Unwrap(err).Error(), nil
	}

	// Error is something else.
	return "System error. Please visit the dashboard.", err
}

// NewVoice will send out the initial Call to Twilio, specifying all details needed for Twilio to make the first call to the end user
// It performs operations like validating essential parameters, registering the Twilio client and db
// and adding routes for successful and unsuccessful call connections to Twilio
func NewVoice(ctx context.Context, db *sql.DB, c *Config) (*Voice, error) {
	v := &Voice{
		c: c,
	}

	return v, nil
}

// SetReceiver sets the notification.Receiver for incoming calls and status updates.
func (v *Voice) SetReceiver(r notification.Receiver) { v.r = r }

func (v *Voice) ServeCall(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	switch CallType(req.FormValue("type")) {
	case CallTypeAlert:
		v.ServeAlert(w, req)
	case CallTypeAlertStatus:
		v.ServeAlertStatus(w, req)
	case CallTypeTest:
		v.ServeTest(w, req)
	case CallTypeStop:
		v.ServeStop(w, req)
	case CallTypeVerify:
		v.ServeVerify(w, req)
	default:
		_, call, _ := v.getCall(w, req)
		if !call.Outbound {
			v.ServeInbound(w, req)
			return
		}
		http.NotFound(w, req)
	}
}

// Status provides the current status of a message.
func (v *Voice) Status(ctx context.Context, externalID string) (*notification.Status, error) {
	call, err := v.c.GetVoice(ctx, externalID)
	if err != nil {
		return nil, err
	}
	return call.messageStatus(), nil
}

// callbackURL returns an absolute URL pointing to the named callback.
// If params is nil, default values from the BaseURL are used.
func (v *Voice) callbackURL(ctx context.Context, params url.Values, typ CallType) string {
	cfg := config.FromContext(ctx)
	p := make(url.Values)
	p.Set("type", string(typ))
	return cfg.CallbackURL("/api/v2/twilio/call", params, p)
}

func spellNumber(n int) string {
	s := strconv.Itoa(n)

	return strings.Join(strings.Split(s, ""), ". ")
}

// Send implements the notification.Sender interface.
func (v *Voice) Send(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	cfg := config.FromContext(ctx)
	if !cfg.Twilio.Enable {
		return nil, errors.New("Twilio provider is disabled")
	}
	toNumber := msg.Destination().Value

	if toNumber == cfg.Twilio.FromNumber {
		return nil, errors.New("refusing to make outgoing call to FromNumber")
	}
	ctx = log.WithFields(ctx, log.Fields{
		"Number": toNumber,
		"Type":   "TwilioVoice",
	})

	opts := &VoiceOptions{
		ValidityPeriod: time.Second * 10,
	}

	if err := opts.setMsgParams(msg); err != nil {
		return nil, err
	}

	msgBody, err := buildMessage(fmt.Sprintf("Hello! This is %s", cfg.ApplicationName()), msg)
	if err != nil {
		return nil, err
	}
	opts.setMsgBody(msgBody)

	voiceResponse, err := v.c.StartVoice(ctx, toNumber, opts)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "call user"))
		return nil, err
	}

	return voiceResponse.sentMessage(), nil
}

func disabled(w http.ResponseWriter, req *http.Request) bool {
	ctx := req.Context()
	cfg := config.FromContext(ctx)
	if !cfg.Twilio.Enable {
		log.Log(ctx, errors.New("Twilio provider is disabled"))
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return true
	}
	return false
}

func (v *Voice) ServeStatusCallback(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}

	ctx := req.Context()
	status := CallStatus(req.FormValue("CallStatus"))
	number := validPhone(req.FormValue("To"))
	sid := validSID(req.FormValue("CallSid"))
	if status == "" || number == "" || sid == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx = log.WithFields(ctx, log.Fields{
		"Status": status,
		"SID":    sid,
		"Phone":  number,
		"Type":   "TwilioVoice",
	})

	if status == CallStatusFailed && req.FormValue("SipResponseCode") == "480" {
		// treat it as no-answer since callee unreachable instead of failed
		status = CallStatusNoAnswer
	}

	callState := &Call{
		SID:    sid,
		Status: status,
		To:     number,
		From:   req.FormValue("From"),
	}
	seq, err := strconv.Atoi(req.FormValue("SequenceNumber"))
	if err == nil {
		callState.SequenceNumber = &seq
	}

	err = v.r.SetMessageStatus(ctx, sid, callState.messageStatus())
	if err != nil {
		// log and continue
		log.Log(ctx, err)
	}
}

type call struct {
	Number     string
	SID        string
	Digits     string
	RetryCount int
	Outbound   bool
	Q          url.Values

	// Embedded query fields
	msgID        string
	msgSubjectID int
	msgBody      string
}

// doDeadline will ensure a return within 5 seconds, and log
// the original error in the event of a timeout.
func doDeadline(ctx context.Context, fn func() error) error {
	errCh := make(chan error, 1)
	timeoutCh := make(chan struct{})
	go func() {
		err := fn()
		errCh <- err
		select {
		case errCh <- nil:
			// error consumed, nothing to do
		case <-timeoutCh:
			// log if error (other than context canceled)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Log(ctx, err)
			}
		}
	}()
	t := time.NewTimer(5 * time.Second)
	defer t.Stop()
	select {
	case err := <-errCh:
		return err
	case <-t.C:
		close(timeoutCh)
		return errVoiceTimeout
	}
}

func (v *Voice) ServeStop(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx, call, errResp := v.getCall(w, req)
	if call == nil {
		return
	}

	resp := newTwiMLResponse(ctx, w)
	switch call.Digits {
	default:
		resp.SayUnknownDigit()
		fallthrough
	case "", digitRepeat:
		resp.AddOptions(optionConfirmStop, optionCancel)
		resp.Gather(v.callbackURL(ctx, call.Q, CallTypeStop))
		return
	case digitConfirm:
		err := doDeadline(ctx, func() error {
			return v.r.Stop(ctx, notification.Dest{Type: notification.DestTypeVoice, Value: call.Number})
		})

		if errResp(false, errors.Wrap(err, "process STOP response"), "") {
			return
		}

		resp.Say("Unenrolled.")
		resp.Hangup()
		return
	case digitGoBack: // Go back to main menu
		resp.Redirect(v.callbackURL(ctx, call.Q, CallType(call.Q.Get("previous"))))
		return
	}
}

type errRespFn func(userErr bool, err error, msg string) bool

func (v *Voice) getCall(w http.ResponseWriter, req *http.Request) (context.Context, *call, errRespFn) {
	ctx := req.Context()
	cfg := config.FromContext(ctx)
	isOutbound := req.FormValue("Direction") == "outbound-api"
	var remoteNumRaw string
	if isOutbound {
		remoteNumRaw = req.FormValue("To")
	} else {
		remoteNumRaw = req.FormValue("From")
	}
	callSID := validSID(req.FormValue("CallSid"))
	phoneNumber := validPhone(remoteNumRaw)
	digits := req.FormValue("Digits")

	if callSID == "" || phoneNumber == "" || phoneNumber == cfg.Twilio.FromNumber {
		http.Error(w, "", http.StatusBadRequest)
		return nil, nil, nil
	}

	q := req.URL.Query()

	retryCount, _ := strconv.Atoi(q.Get("retry_count"))
	q.Del("retry_count") // retry_count will only be set again if we go through the errResp

	msgID := q.Get(msgParamID)
	subID, _ := strconv.Atoi(q.Get(msgParamSubID))
	bodyData, _ := b64enc.DecodeString(q.Get(msgParamBody))

	if isOutbound && msgID == "" {
		log.Log(ctx, errors.Errorf("parse call: query param %s is empty or invalid", msgParamID))
	}
	if isOutbound && subID == 0 {
		log.Log(ctx, errors.Errorf("parse call: query param %s is empty or invalid", msgParamSubID))
	}
	if isOutbound && len(bodyData) == 0 {
		log.Log(ctx, errors.Errorf("parse call: query param %s is empty or invalid", msgParamBody))
	}

	if digits == "" {
		digits = q.Get("retry_digits")
	}
	q.Del("retry_digits")

	ctx = log.WithFields(ctx, log.Fields{
		"SID":    callSID,
		"Phone":  phoneNumber,
		"Digits": digits,
		"Type":   "TwilioVoice",
	})

	errResp := func(userErr bool, err error, msg string) bool {
		if err == nil {
			return false
		}

		// always log the failure
		log.Log(ctx, err)

		if (errors.Is(err, errVoiceTimeout) || retry.IsTemporaryError(err)) && retryCount < 3 {
			// schedule a retry
			q.Set("retry_count", strconv.Itoa(retryCount+1))
			q.Set("retry_digits", digits)

			newTwiMLResponse(ctx, w).
				Say("One moment please.").
				RedirectPauseSec(v.callbackURL(ctx, q, CallType(q.Get("type"))), 5)

			return true
		}

		newTwiMLResponse(ctx, w).Say("An error has occurred. Please use the dashboard to manage alerts.").Hangup()
		return true
	}

	return ctx, &call{
		Number:     phoneNumber,
		SID:        callSID,
		RetryCount: retryCount,
		Digits:     digits,
		Outbound:   isOutbound,
		Q:          q,

		msgID:        msgID,
		msgSubjectID: subID,
		msgBody:      string(bodyData),
	}, errResp
}

func (v *Voice) ServeTest(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx, call, _ := v.getCall(w, req)
	if call == nil {
		return
	}

	resp := newTwiMLResponse(ctx, w)
	switch call.Digits {
	default:
		resp.SayUnknownDigit()
		fallthrough
	case "", digitRepeat:
		resp.Say(call.msgBody)
		resp.AddOptions(optionStop)
		resp.Gather(v.callbackURL(ctx, call.Q, CallTypeTest))
		return
	case digitStop:
		call.Q.Set("previous", string(CallTypeTest))
		resp.Redirect(v.callbackURL(ctx, call.Q, CallTypeStop))
		return
	}
}

func (v *Voice) ServeVerify(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx, call, _ := v.getCall(w, req)
	if call == nil {
		return
	}

	resp := newTwiMLResponse(ctx, w)
	switch call.Digits {
	default:
		resp.SayUnknownDigit()
		fallthrough
	case "", digitRepeat:
		resp.Say(call.msgBody)
		resp.Gather(v.callbackURL(ctx, call.Q, CallTypeVerify))
		return
	}
}

func (v *Voice) ServeAlertStatus(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx, call, _ := v.getCall(w, req)
	if call == nil {
		return
	}

	resp := newTwiMLResponse(ctx, w)
	switch call.Digits {
	default:
		resp.SayUnknownDigit()
		fallthrough
	case "", digitRepeat:
		resp.Say(call.msgBody)
		resp.AddOptions(optionStop)
		resp.Gather(v.callbackURL(ctx, call.Q, CallTypeAlertStatus))
		return
	case digitStop:
		call.Q.Set("previous", string(CallTypeAlertStatus))
		resp.Redirect(v.callbackURL(ctx, call.Q, CallTypeStop))
		return
	}
}

// ServeInbound is the handler for inbound calls.
func (v *Voice) ServeInbound(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx, call, _ := v.getCall(w, req)
	if call == nil {
		return
	}
	cfg := config.FromContext(ctx)

	resp := newTwiMLResponse(ctx, w)
	switch call.Digits {
	default:
		resp.SayUnknownDigit()
		fallthrough
	case "", digitRepeat:
		resp.Sayf("Hello! This is %s. ", cfg.ApplicationName())
		resp.Say("Please use the application dashboard to manage alerts.")
		resp.AddOptions(optionStop)
		resp.Gather(v.callbackURL(ctx, call.Q, ""))
		return
	case digitStop:
		call.Q.Set("previous", "")
		resp.Redirect(v.callbackURL(ctx, call.Q, CallTypeStop))
		return
	}
}

// ServeAlert serves a call for an alert notification.
func (v *Voice) ServeAlert(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx, call, errResp := v.getCall(w, req)
	if call == nil {
		return
	}

	// See Twilio Request Parameter documentation at
	// https://www.twilio.com/docs/api/twiml/twilio_request#synchronous
	resp := newTwiMLResponse(ctx, w)
	switch call.Digits {
	default:
		if call.Digits == digitOldAck {
			resp.Sayf("The menu options have changed. To acknowledge, press %s.", digitAck)
		} else if call.Digits == digitOldClose {
			resp.Sayf("The menu options have changed. To close, press %s.", digitClose)
		} else {
			resp.SayUnknownDigit()
		}
		fallthrough
	case "", digitRepeat:
		resp.Say(call.msgBody)
		if call.Q.Get(msgParamBundle) == "1" {
			resp.AddOptions(optionAckAll, optionCloseAll)
		} else {
			resp.AddOptions(optionAck, optionEscalate, optionClose)
		}
		resp.AddOptions(optionStop)
		resp.Gather(v.callbackURL(ctx, call.Q, CallTypeAlert))
		return

	case digitStop:
		call.Q.Set("previous", string(CallTypeAlert))
		resp.Redirect(v.callbackURL(ctx, call.Q, CallTypeStop))
		return

	case digitAck, digitClose, digitEscalate: // Acknowledge , Escalate and Close cases
		var result notification.Result
		var msg string
		if call.Digits == digitClose {
			result = notification.ResultResolve
			msg = "Closed"
		} else if call.Digits == digitEscalate {
			result = notification.ResultEscalate
			msg = "Escalation requested"
		} else {
			result = notification.ResultAcknowledge
			msg = "Acknowledged"
		}
		if call.Q.Get(msgParamBundle) == "1" {
			msg += " all alerts."
		}
		err := doDeadline(ctx, func() error {
			return v.r.Receive(ctx, call.msgID, result)
		})
		if err != nil {
			msg, err = voiceErrorMessage(ctx, err)
		}
		if errResp(false, errors.Wrap(err, "process response"), "Failed to process notification response.") {
			return
		}

		resp.Say(msg).Hangup()
		return
	}
}

// FriendlyValue will return the international formatting of the phone number.
func (v *Voice) FriendlyValue(ctx context.Context, value string) (string, error) {
	num, err := phonenumbers.Parse(value, "")
	if err != nil {
		return "", fmt.Errorf("parse number for formatting: %w", err)
	}
	return phonenumbers.Format(num, phonenumbers.INTERNATIONAL), nil
}

// buildMessage is a function that will build the VoiceOptions object with the proper message contents
func buildMessage(prefix string, msg notification.Message) (message string, err error) {
	if prefix == "" {
		return "", errors.New("buildMessage error: no prefix provided")
	}

	switch t := msg.(type) {
	case notification.AlertBundle:
		message = fmt.Sprintf("%s with alert notifications. Service '%s' has %d unacknowledged alerts.", prefix, t.ServiceName, t.Count)
	case notification.Alert:
		if t.Summary == "" {
			t.Summary = "No summary provided"
		}
		message = fmt.Sprintf("%s with an alert notification. %s.", prefix, t.Summary)
	case notification.AlertStatus:
		message = rmParen.ReplaceAllString(t.LogEntry, "")
		message = fmt.Sprintf("%s with a status update for alert '%s'. %s", prefix, t.Summary, message)
	case notification.Test:
		message = fmt.Sprintf("%s with a test message.", prefix)
	case notification.Verification:
		count := int(math.Log10(float64(t.Code)) + 1)
		message = fmt.Sprintf(
			"%s with your %d-digit verification code. The code is: %s. Again, your %d-digit verification code is: %s.",
			prefix, count, spellNumber(t.Code), count, spellNumber(t.Code),
		)
	default:
		return "", errors.Errorf("unhandled message type: %T", t)
	}

	return
}
