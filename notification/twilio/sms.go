package twilio

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
)

var (
	lastReplyRx  = regexp.MustCompile(`^'?\s*(c|close|a|ack[a-z]*)\s*'?$`)
	shortReplyRx = regexp.MustCompile(`^'?\s*([0-9]+)\s*(c|a)\s*'?$`)
	alertReplyRx = regexp.MustCompile(`^'?\s*(c|close|a|ack[a-z]*)\s*#?\s*([0-9]+)\s*'?$`)

	svcReplyRx = regexp.MustCompile(`^'?\s*([0-9]+)\s*(cc|aa)\s*'?$`)
)

// SMS implements a notification.Sender for Twilio SMS.
type SMS struct {
	b *dbSMS
	c *Config
	r notification.Receiver

	ban *dbBan
}

var _ notification.ReceiverSetter = &SMS{}
var _ notification.Sender = &SMS{}

// NewSMS performs operations like validating essential parameters, registering the Twilio client and db
// and adding routes for successful and unsuccessful message delivery to Twilio
func NewSMS(ctx context.Context, db *sql.DB, c *Config) (*SMS, error) {
	b, err := newDB(ctx, db)
	if err != nil {
		return nil, err
	}

	s := &SMS{
		b: b,
		c: c,
	}
	s.ban, err = newBanDB(ctx, db, c, "twilio_sms_errors")
	if err != nil {
		return nil, errors.Wrap(err, "init Twilio SMS DB")
	}

	return s, nil
}

// SetReceiver sets the notification.Receiver for incoming messages and status updates.
func (s *SMS) SetReceiver(r notification.Receiver) { s.r = r }

// Status provides the current status of a message.
func (s *SMS) Status(ctx context.Context, id, providerID string) (*notification.MessageStatus, error) {
	msg, err := s.c.GetSMS(ctx, providerID)
	if err != nil {
		return nil, err
	}
	return msg.messageStatus(id), nil
}

// Send implements the notification.Sender interface.
func (s *SMS) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {
	cfg := config.FromContext(ctx)
	if !cfg.Twilio.Enable {
		return nil, errors.New("Twilio provider is disabled")
	}
	if msg.Destination().Type != notification.DestTypeSMS {
		return nil, errors.Errorf("unsupported destination type %s; expected SMS", msg.Destination().Type)
	}
	destNumber := msg.Destination().Value
	if destNumber == cfg.Twilio.FromNumber {
		return nil, errors.New("refusing to send outgoing SMS to FromNumber")
	}

	ctx = log.WithFields(ctx, log.Fields{
		"Phone": destNumber,
		"Type":  "TwilioSMS",
	})

	b, err := s.ban.IsBanned(ctx, destNumber, true)
	if err != nil {
		return nil, errors.Wrap(err, "check ban status")
	}
	if b {
		return nil, errors.New("number had too many outgoing errors recently")
	}

	makeSMSCode := func(alertID int, serviceID string) int {
		var code int
		if hasTwoWaySMSSupport(ctx, destNumber) {
			code, err = s.b.insertDB(ctx, destNumber, msg.ID(), alertID, serviceID)
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "insert alert id for SMS callback -- sending 1-way SMS as fallback"))
			}
		}
		return code
	}

	var message string
	switch t := msg.(type) {
	case notification.AlertStatus:
		message, err = alertSMS{
			ID:   t.AlertID,
			Body: t.LogEntry,
		}.Render()
	case notification.AlertStatusBundle:
		message, err = alertSMS{
			ID:    t.AlertID,
			Body:  t.LogEntry,
			Count: t.Count - 1,
		}.Render()
	case notification.AlertBundle:
		var link string
		if !cfg.General.DisableSMSLinks {
			link = cfg.CallbackURL(fmt.Sprintf("/services/%s/alerts", t.ServiceID))
		}

		message, err = alertSMS{
			Count: t.Count,
			Body:  t.ServiceName,
			Link:  link,
			Code:  makeSMSCode(0, t.ServiceID),
		}.Render()
	case notification.Alert:
		var link string
		if !cfg.General.DisableSMSLinks {
			link = cfg.CallbackURL(fmt.Sprintf("/alerts/%d", t.AlertID))
		}

		message, err = alertSMS{
			ID:   t.AlertID,
			Body: t.Summary,
			Link: link,
			Code: makeSMSCode(t.AlertID, ""),
		}.Render()
	case notification.Test:
		message = "This is a test message from GoAlert."
	case notification.Verification:
		message = fmt.Sprintf("GoAlert verification code: %d", t.Code)
	default:
		return nil, errors.Errorf("unhandled message type %T", t)
	}
	if err != nil {
		return nil, errors.Wrap(err, "render message")
	}

	opts := &SMSOptions{
		ValidityPeriod: time.Second * 10,
		CallbackParams: make(url.Values),
	}
	opts.CallbackParams.Set(msgParamID, msg.ID())
	// Actually send notification to end user & receive Message Status
	resp, err := s.c.SendSMS(ctx, destNumber, message, opts)
	if err != nil {
		return nil, errors.Wrap(err, "send message")
	}

	return resp.messageStatus(msg.ID()), nil
}

func (s *SMS) ServeStatusCallback(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx := req.Context()
	status := MessageStatus(req.FormValue("MessageStatus"))
	sid := validSID(req.FormValue("MessageSid"))
	number := validPhone(req.FormValue("To"))
	if status == "" || sid == "" || number == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx = log.WithFields(ctx, log.Fields{
		"Status": status,
		"SID":    sid,
		"Phone":  number,
		"Type":   "TwilioSMS",
	})
	msg := Message{SID: sid, Status: status}

	log.Debugf(ctx, "Got Twilio SMS status callback.")

	err := s.r.UpdateStatus(ctx, msg.messageStatus(req.URL.Query().Get(msgParamID)))
	if err != nil {
		// log and continue
		log.Log(ctx, err)
	}

	if status != MessageStatusFailed {
		// ignore other types
		return
	}

	err = s.ban.RecordError(context.Background(), number, true, "send failed")
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "record error"))
	}

}

// isStopMessage checks the body of the message against single-word matches
// i.e. "stop" will unsubscribe, however "please stop" will not.
func isStopMessage(body string) bool {
	switch strings.ToLower(body) {
	case "stop", "stopall", "unsubscribe", "cancel", "end", "quit":
		return true
	}

	return false
}

// isStartMessage checks the body of the message against single-word matches
// i.e. "start" will resubscribe, however "please start" will not.
func isStartMessage(body string) bool {
	switch strings.ToLower(body) {
	case "start", "yes", "unstop":
		return true
	}

	return false
}

func (s *SMS) ServeMessage(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx := req.Context()
	cfg := config.FromContext(ctx)
	from := validPhone(req.FormValue("From"))
	if from == "" || from == cfg.Twilio.FromNumber {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx = log.WithFields(ctx, log.Fields{
		"Number": from,
		"Type":   "TwilioSMS",
	})

	respond := func(errMsg string, msg string) {
		if errMsg != "" {
			err := s.ban.RecordError(context.Background(), from, false, errMsg)
			if err != nil {
				log.Log(ctx, errors.Wrap(err, "record error"))
			}
		}
		_, err := s.c.SendSMS(ctx, from, msg, &SMSOptions{FromNumber: req.FormValue("to")})
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "send response"))
		}
		// TODO: we should track & queue these
		// (maybe the engine should generate responses instead)
	}
	var banned bool
	var err error
	retryOpts := []retry.Option{
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Second),
	}

	err = retry.DoTemporaryError(func(int) error {
		banned, err = s.ban.IsBanned(ctx, from, false)
		return errors.Wrap(err, "look up ban status")
	}, retryOpts...)
	if err != nil {
		log.Log(ctx, err)
		respond("", "System error. Visit the dashboard to manage alerts.")
		return
	}
	if banned {
		http.Error(w, "", http.StatusTooManyRequests)
		return
	}

	// handle start and stop codes from user
	body := req.FormValue("Body")
	dest := notification.Dest{Type: notification.DestTypeSMS, Value: from}
	if isStartMessage(body) {
		err := retry.DoTemporaryError(func(int) error { return s.r.Start(ctx, dest) }, retryOpts...)
		if err != nil {
			log.Log(ctx, fmt.Errorf("process START message: %w", err))
		}
		return
	}
	if isStopMessage(body) {
		err := retry.DoTemporaryError(func(int) error { return s.r.Stop(ctx, dest) }, retryOpts...)
		if err != nil {
			log.Log(ctx, fmt.Errorf("process STOP message: %w", err))
		}
		return
	}

	if cfg.Twilio.DisableTwoWaySMS {
		respond("response codes disabled", "Response codes are currently disabled. Visit the dashboard to manage alerts.")
		return
	}

	body = strings.TrimSpace(body)
	body = strings.ToLower(body)
	var lookupFn func() (*codeInfo, error)
	var result notification.Result
	var isSvc bool
	if m := lastReplyRx.FindStringSubmatch(body); len(m) == 2 {
		if strings.HasPrefix(m[1], "a") {
			result = notification.ResultAcknowledge
		} else {
			result = notification.ResultResolve
		}
		lookupFn = func() (*codeInfo, error) { return s.b.LookupByCode(ctx, from, 0) }
	} else if m := shortReplyRx.FindStringSubmatch(body); len(m) == 3 {
		if strings.HasPrefix(m[2], "a") {
			result = notification.ResultAcknowledge
		} else {
			result = notification.ResultResolve
		}
		code, err := strconv.Atoi(m[1])
		if err != nil {
			log.Debug(ctx, errors.Wrap(err, "parse code"))
		} else {
			ctx = log.WithField(ctx, "Code", code)
			lookupFn = func() (*codeInfo, error) { return s.b.LookupByCode(ctx, from, code) }
		}
	} else if m := alertReplyRx.FindStringSubmatch(body); len(m) == 3 {
		if strings.HasPrefix(m[1], "a") {
			result = notification.ResultAcknowledge
		} else {
			result = notification.ResultResolve
		}
		alertID, err := strconv.Atoi(m[2])
		if err != nil {
			log.Debug(ctx, errors.Wrap(err, "parse alertID"))
		} else {
			ctx = log.WithField(ctx, "AlertID", alertID)
			lookupFn = func() (*codeInfo, error) { return s.b.LookupByAlertID(ctx, from, alertID) }
		}
	} else if m := svcReplyRx.FindStringSubmatch(body); len(m) == 3 {
		isSvc = true
		if strings.HasPrefix(m[2], "a") {
			result = notification.ResultAcknowledge
		} else {
			result = notification.ResultResolve
		}
		code, err := strconv.Atoi(m[1])
		if err != nil {
			log.Debug(ctx, errors.Wrap(err, "parse code"))
		} else {
			ctx = log.WithField(ctx, "Code", code)
			lookupFn = func() (*codeInfo, error) { return s.b.LookupSvcByCode(ctx, from, code) }
		}
	}

	if lookupFn == nil {
		respond("unknown action", "Sorry, but that isn't a request GoAlert understood. Visit the Web UI for more information. To unsubscribe, reply with STOP.")
		ctx = log.WithField(ctx, "SMSBody", body)
		log.Debug(ctx, errors.Wrap(err, "parse alert action"))
		return
	}

	var prefix string
	if result == notification.ResultAcknowledge {
		prefix = "Acknowledged"
	} else {
		prefix = "Closed"
	}

	var nonSystemErr bool
	var info *codeInfo
	err = retry.DoTemporaryError(func(int) error {
		info, err = lookupFn()
		if err != nil {
			return errors.Wrap(err, "lookup code")
		}

		err = s.r.Receive(ctx, info.CallbackID, result)
		if err != nil {
			return fmt.Errorf("process notification response: %w", err)
		}
		return nil
	}, retryOpts...)

	if errors.Is(err, sql.ErrNoRows) || (isSvc && info.ServiceName == "") || (!isSvc && info.AlertID == 0) {
		respond("unknown callbackID", "Unknown reply code for this action. Visit the dashboard to manage alerts.")
		return
	}

	msg := "System error. Visit the dashboard to manage alerts."
	if alert.IsAlreadyClosed(err) {
		nonSystemErr = true
		msg = fmt.Sprintf("Alert #%d already closed", alert.AlertID(err))
	} else if alert.IsAlreadyAcknowledged(err) {
		nonSystemErr = true
		msg = fmt.Sprintf("Alert #%d already acknowledged", alert.AlertID(err))
	}

	if nonSystemErr {
		var e alert.LogEntryFetcher
		// alert store returns the special error struct, twilio checks if it's special, and if so, pulls the log entry
		if errors.As(err, &e) {
			err = nil
			// we pass a 'sudo' context to give permission
			permission.SudoContext(ctx, func(sCtx context.Context) {
				entry, err := e.LogEntry(sCtx)
				if err != nil {
					log.Log(sCtx, errors.Wrap(err, "fetch log entry"))
				} else {
					msg += "\n\n" + entry.String()
				}
			})
		}
		log.Log(ctx, errors.Wrap(err, "process notification response"))
		respond("", msg)
		return
	}

	if err != nil {
		log.Log(ctx, err)
		respond("", msg)
		return
	}

	if info.ServiceName != "" {
		respond("", fmt.Sprintf("%s all alerts for service '%s'", prefix, info.ServiceName))
	} else {
		respond("", fmt.Sprintf("%s alert #%d", prefix, info.AlertID))
	}
}
