package twilio

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
)

// CallType indicates a supported Twilio voice call type.
type CallType string

// Supported call types.
const (
	CallTypeAlert       = CallType("alert")
	CallTypeAlertStatus = CallType("alert-status")
	CallTypeTest        = CallType("test")
	CallTypeVerify      = CallType("verify")
	CallTypeStop        = CallType("stop")
)

// We use url encoding with no padding to try and eliminate
// encoding problems with buggy apps.
var b64enc = base64.URLEncoding.WithPadding(base64.NoPadding)

var errVoiceTimeout = errors.New("process voice action: timeout")

// KeyPressed specifies a key pressed from the voice menu options.
type KeyPressed string

// Possible keys pressed from the Menu mapped to their actions.
const (
	digitAck      = "4"
	digitClose    = "6"
	digitStop     = "1"
	digitGoBack   = "1"
	digitRepeat   = "*"
	digitConfirm  = "3"
	digitOldAck   = "8"
	digitOldClose = "9"
)

var pRx = regexp.MustCompile(`\((.*?)\)`)

// Voice implements a notification.Sender for Twilio voice calls.
type Voice struct {
	c   *Config
	ban *dbBan

	respCh chan *notification.MessageResponse
	statCh chan *notification.MessageStatus
}

type gather struct {
	XMLName   xml.Name `xml:"Gather,omitempty"`
	Action    string   `xml:"action,attr,omitempty"`
	Method    string   `xml:"method,attr,omitempty"`
	NumDigits int      `xml:"numDigits,attr,omitempty"`
	Say       string   `xml:"Say,omitempty"`
}

type twiMLRedirect struct {
	XMLName     xml.Name `xml:"Response"`
	RedirectURL string   `xml:"Redirect"`
}

type twiMLRetry struct {
	XMLName xml.Name `xml:"Response"`
	Say     string   `xml:"Say"`
	Pause   struct {
		Seconds int `xml:"length,attr"`
	} `xml:"Pause"`
	RedirectURL string `xml:"Redirect"`
}

type twiMLGather struct {
	XMLName xml.Name `xml:"Response"`
	Gather  *gather
}
type twiMLEnd struct {
	XMLName xml.Name `xml:"Response"`
	Say     string   `xml:"Say,omitempty"`
	Hangup  struct{}
}

var rmParen = regexp.MustCompile(`\s*\(.*?\)`)

func voiceErrorMessage(ctx context.Context, err error) (string, error) {
	if e, ok := errors.Cause(err).(alert.LogEntryFetcher); ok {
		// we pass a 'sudo' context to give permission
		var msg string
		permission.SudoContext(ctx, func(sCtx context.Context) {
			entry, err := e.LogEntry(sCtx)
			if err != nil {
				log.Log(sCtx, errors.Wrap(err, "fetch log entry"))
			} else {
				// Stripping off anything in between parenthesis
				msg = "Already " + pRx.ReplaceAllString(entry.String(), "")
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
	// Error is something else.
	return "System error. Please visit the dashboard.", err
}

// NewVoice will send out the initial Call to Twilio, specifying all details needed for Twilio to make the first call to the end user
// It performs operations like validating essential parameters, registering the Twilio client and db
// and adding routes for successful and unsuccessful call connections to Twilio
func NewVoice(ctx context.Context, db *sql.DB, c *Config) (*Voice, error) {
	v := &Voice{
		c: c,

		respCh: make(chan *notification.MessageResponse),
		statCh: make(chan *notification.MessageStatus, 10),
	}

	var err error
	v.ban, err = newBanDB(ctx, db, c, "twilio_voice_errors")
	if err != nil {
		return nil, errors.Wrap(err, "init voice ban DB")
	}

	return v, nil
}

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
		http.NotFound(w, req)
	}
}

// Status provides the current status of a message.
func (v *Voice) Status(ctx context.Context, id, providerID string) (*notification.MessageStatus, error) {
	call, err := v.c.GetVoice(ctx, providerID)
	if err != nil {
		return nil, err
	}
	return call.messageStatus(id), nil
}

// ListenStatus will return a channel that is fed async status updates.
func (v *Voice) ListenStatus() <-chan *notification.MessageStatus { return v.statCh }

// ListenResponse will return a channel that is fed async message responses.
func (v *Voice) ListenResponse() <-chan *notification.MessageResponse { return v.respCh }

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

	return strings.Join(strings.Split(s, ""), ", ")
}

// Send implements the notification.Sender interface.
func (v *Voice) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {
	cfg := config.FromContext(ctx)
	if !cfg.Twilio.Enable {
		return nil, errors.New("Twilio provider is disabled")
	}
	toNumber := msg.Destination().Value
	if !supportedCountryCode(toNumber) {
		return nil, errors.New("unsupported country code")
	}

	if toNumber == cfg.Twilio.FromNumber {
		return nil, errors.New("refusing to make outgoing call to FromNumber")
	}
	ctx = log.WithFields(ctx, log.Fields{
		"Number": toNumber,
		"Type":   "TwilioVoice",
	})
	b, err := v.ban.IsBanned(ctx, toNumber, true)
	if err != nil {
		return nil, errors.Wrap(err, "check ban status")
	}
	if b {
		return nil, errors.New("number had too many outgoing errors recently")
	}

	var ep CallType
	var message string

	switch msg.Type() {
	case notification.MessageTypeAlert:
		message = msg.Body()
		ep = CallTypeAlert
	case notification.MessageTypeAlertStatus:
		message = rmParen.ReplaceAllString(msg.Body(), "")
		ep = CallTypeAlertStatus
	case notification.MessageTypeTest:
		message = "This is a test message from GoAlert."
		ep = CallTypeTest
	case notification.MessageTypeVerification:
		message = "Your verification code for GoAlert is: " + spellNumber(msg.SubjectID())
		ep = CallTypeVerify
	default:
		return nil, errors.Errorf("unhandled message type %s", msg.Type().String())
	}

	if message == "" {
		message = "No summary provided."
	}

	opts := &VoiceOptions{
		ValidityPeriod: time.Second * 10,
		CallType:       ep,
		CallbackParams: make(url.Values),
		Params:         make(url.Values),
	}
	opts.CallbackParams.Set(msgParamID, msg.ID())
	opts.Params.Set(msgParamSubID, strconv.Itoa(msg.SubjectID()))
	// Encode the body so we don't need to worry about
	// buggy apps not escaping url params properly.
	opts.Params.Set(msgParamBody, b64enc.EncodeToString([]byte(message)))

	voiceResponse, err := v.c.StartVoice(ctx, toNumber, opts)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "call user"))
		return nil, err
	}

	return voiceResponse.messageStatus(msg.ID()), nil
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
	}
	seq, err := strconv.Atoi(req.FormValue("SequenceNumber"))
	if err == nil {
		callState.SequenceNumber = &seq
	}

	v.statCh <- callState.messageStatus(req.URL.Query().Get(msgParamID))

	// Only update current call we are on, except for failed call
	if status != CallStatusFailed {
		return
	}

	err = v.ban.RecordError(context.Background(), number, true, "send failed")
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "record error"))
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

func (v *Voice) handleResponse(resp notification.MessageResponse) error {
	errCh := make(chan error, 1)
	resp.Err = errCh
	v.respCh <- &resp
	var err error
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	select {
	case err = <-errCh:
	case <-t.C:
		err = errVoiceTimeout
	}
	return err
}

func (v *Voice) ServeStop(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx, call, errResp := v.getCall(w, req)
	if call == nil {
		return
	}

	var messagePrefix string
	switch call.Digits {
	default:
		messagePrefix = "I am sorry. I didn't catch that. "
		fallthrough

	case "", digitRepeat:
		message := fmt.Sprintf("%sTo confirm unenrollment of this number, press %s. To go back to main menu, press %s. To repeat this message, press %s.", messagePrefix, digitConfirm, digitGoBack, digitRepeat)
		g := &gather{
			Action:    v.callbackURL(ctx, call.Q, CallTypeStop),
			Method:    "POST",
			NumDigits: 1,
			Say:       message,
		}
		renderXML(w, req, twiMLGather{
			Gather: g,
		})

	case digitConfirm:
		err := v.handleResponse(notification.MessageResponse{
			Ctx:    ctx,
			From:   notification.Dest{Type: notification.DestTypeVoice, Value: call.Number},
			Result: notification.ResultStop,
		})

		if errResp(false, errors.Wrap(err, "process STOP response"), "") {
			return
		}

		renderXML(w, req, twiMLEnd{
			Say: "Unenrolled. Goodbye.",
		})
	case digitGoBack: // Go back to main menu
		renderXML(w, req, twiMLRedirect{
			RedirectURL: v.callbackURL(ctx, call.Q, CallType(call.Q.Get("previous"))),
		})
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

	if !isOutbound {
		return ctx, &call{
			Number:     phoneNumber,
			SID:        callSID,
			RetryCount: retryCount,
			Digits:     digits,
			Outbound:   isOutbound,
			Q:          q,
		}, nil
	}

	msgID := q.Get(msgParamID)
	subID, _ := strconv.Atoi(q.Get(msgParamSubID))
	bodyData, _ := b64enc.DecodeString(q.Get(msgParamBody))
	if msgID == "" {
		log.Log(ctx, errors.Errorf("parse call: query param %s is empty or invalid", msgParamID))
	}
	if subID == 0 {
		log.Log(ctx, errors.Errorf("parse call: query param %s is empty or invalid", msgParamSubID))
	}
	if len(bodyData) == 0 {
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
		if userErr {
			rerr := v.ban.RecordError(context.Background(), phoneNumber, isOutbound, msg)
			if rerr != nil {
				log.Log(ctx, errors.Wrap(rerr, "record error"))
			}
		}

		// always log the failure
		log.Log(ctx, err)

		if (errors.Cause(err) == errVoiceTimeout || retry.IsTemporaryError(err)) && retryCount < 3 {
			// schedule a retry
			q.Set("retry_count", strconv.Itoa(retryCount+1))
			q.Set("retry_digits", digits)
			retry := twiMLRetry{
				Say:         "One moment please.",
				RedirectURL: v.callbackURL(ctx, q, CallType(q.Get("type"))),
			}
			retry.Pause.Seconds = 5
			renderXML(w, req, retry)
			return true
		}

		renderXML(w, req, twiMLEnd{
			Say: "An error has occurred. Please login to the dashboard to manage alerts. Goodbye.",
		})
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

func renderXML(w http.ResponseWriter, req *http.Request, v interface{}) {
	data, err := xml.Marshal(v)
	if errutil.HTTPError(req.Context(), w, err) {
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	io.WriteString(w, xml.Header)
	w.Write(data)
}

func (v *Voice) ServeTest(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx, call, _ := v.getCall(w, req)
	if call == nil {
		return
	}

	var messagePrefix string
	switch call.Digits {
	default:
		messagePrefix = "I am sorry. I didn't catch that. "
		fallthrough
	case "", digitRepeat:
		message := fmt.Sprintf("%s%s. To repeat this message, press %s.", messagePrefix, call.msgBody, digitRepeat)
		g := &gather{
			Action:    v.callbackURL(ctx, call.Q, CallTypeTest),
			Method:    "POST",
			NumDigits: 1,
			Say:       message,
		}
		renderXML(w, req, twiMLGather{
			Gather: g,
		})
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

	var messagePrefix string
	switch call.Digits {
	default:
		messagePrefix = "I am sorry. I didn't catch that. "
		fallthrough
	case "", digitRepeat:
		message := fmt.Sprintf("%s%s. To repeat this message, press %s.", messagePrefix, call.msgBody, digitRepeat)
		g := &gather{
			Action:    v.callbackURL(ctx, call.Q, CallTypeVerify),
			Method:    "POST",
			NumDigits: 1,
			Say:       message,
		}
		renderXML(w, req, twiMLGather{
			Gather: g,
		})
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

	var messagePrefix string
	switch call.Digits {
	default:
		messagePrefix = "I am sorry. I didn't catch that. "
		fallthrough
	case "", digitRepeat:
		message := fmt.Sprintf(
			"%sStatus update for Alert number %d. %s. To repeat this message, press %s. To unenroll from all notifications, press %s. If you are done, you may simply hang up.",
			messagePrefix, call.msgSubjectID, call.msgBody, digitRepeat, digitStop)
		g := &gather{
			Action:    v.callbackURL(ctx, call.Q, CallTypeAlertStatus),
			Method:    "POST",
			NumDigits: 1,
			Say:       message,
		}
		renderXML(w, req, twiMLGather{
			Gather: g,
		})
		return
	case digitStop:
		call.Q.Set("previous", string(CallTypeAlertStatus))
		renderXML(w, req, twiMLRedirect{
			RedirectURL: v.callbackURL(ctx, call.Q, CallTypeStop),
		})
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

	var err error
	if !call.Outbound {
		// v.serveVoiceUI(w, req)
		err = v.ban.RecordError(context.Background(), call.Number, false, "incoming calls not supported")
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "record error"))
		}
		renderXML(w, req, twiMLEnd{
			Say: "Please login to the dashboard to manage alerts. Goodbye.",
		})
		return
	}

	// See Twilio Request Parameter documentation at
	// https://www.twilio.com/docs/api/twiml/twilio_request#synchronous
	var messagePrefix string

	switch call.Digits {
	default:
		if call.Digits == digitOldAck {
			messagePrefix = fmt.Sprintf("The menu options have changed. To acknowledge, press %s. ", digitAck)
		} else if call.Digits == digitOldClose {
			messagePrefix = fmt.Sprintf("The menu options have changed. To close, press %s. ", digitClose)
		} else {
			messagePrefix = "I am sorry. I didn't catch that. "
		}
		fallthrough
	case "", digitRepeat:
		message := fmt.Sprintf(
			"%sMessage from Go Alert. %s. To acknowledge, press %s. To close, press %s. To unenroll from all notifications, press %s. To repeat this message, press %s",
			messagePrefix, call.msgBody, digitAck, digitClose, digitStop, digitRepeat)
		// User wants Twilio to repeat the message
		g := &gather{
			Action:    v.callbackURL(ctx, call.Q, CallTypeAlert),
			Method:    "POST",
			NumDigits: 1,
			Say:       message,
		}
		renderXML(w, req, twiMLGather{
			Gather: g,
		})
		return

	case digitStop:
		call.Q.Set("previous", string(CallTypeAlert))
		renderXML(w, req, twiMLRedirect{
			RedirectURL: v.callbackURL(ctx, call.Q, CallTypeStop),
		})
		return

	case digitAck, digitClose: // Acknowledge and Close cases
		var result notification.Result
		var msg string
		if call.Digits == digitClose {
			result = notification.ResultResolve
			msg = "Closed. Goodbye."
		} else {
			result = notification.ResultAcknowledge
			msg = "Acknowledged. Goodbye."
		}
		err := v.handleResponse(notification.MessageResponse{
			Ctx:    ctx,
			ID:     call.msgID,
			From:   notification.Dest{Type: notification.DestTypeVoice, Value: call.Number},
			Result: result,
		})
		if err != nil {
			msg, err = voiceErrorMessage(ctx, err)
		}
		if errResp(false, errors.Wrap(err, "process response"), "Failed to process notification response.") {
			return
		}

		renderXML(w, req, twiMLEnd{
			Say: msg,
		})
		return
	}
}
