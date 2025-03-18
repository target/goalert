package twilio

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"

	"github.com/pkg/errors"
)

var (
	lastReplyRx  = regexp.MustCompile(`^'?\s*(c|close|a|e|ack[a-z]*)\s*'?$`)
	shortReplyRx = regexp.MustCompile(`^'?\s*([0-9]+)\s*(c|a|e)\s*'?$`)
	alertReplyRx = regexp.MustCompile(`^'?\s*(c|close|e|a|ack[a-z]*)\s*#?\s*([0-9]+)\s*'?$`)

	svcReplyRx = regexp.MustCompile(`^'?\s*([0-9]+)\s*(cc|aa)\s*'?$`)
)

func NewSMSDest(number string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeTwilioSMS, FieldPhoneNumber, number)
}

// SMS implements a notification.Sender for Twilio SMS.
type SMS struct {
	b *dbSMS
	c *Config
	r notification.Receiver

	limit *replyLimiter
}

var (
	_ notification.ReceiverSetter = &SMS{}
	_ nfydest.MessageSender       = &SMS{}
	_ nfydest.MessageStatuser     = &SMS{}
)

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

		limit: newReplyLimiter(),
	}

	return s, nil
}

// SetReceiver sets the notification.Receiver for incoming messages and status updates.
func (s *SMS) SetReceiver(r notification.Receiver) { s.r = r }

// Status provides the current status of a message.
func (s *SMS) MessageStatus(ctx context.Context, externalID string) (*notification.Status, error) {
	msg, err := s.c.GetSMS(ctx, externalID)
	if err != nil {
		return nil, err
	}

	return msg.messageStatus(), nil
}

// Send implements the notification.Sender interface.
func (s *SMS) SendMessage(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	cfg := config.FromContext(ctx)
	if !cfg.Twilio.Enable {
		return nil, errors.New("Twilio provider is disabled")
	}
	if msg.DestType() != DestTypeTwilioSMS {
		return nil, errors.Errorf("unsupported destination type %s; expected SMS", msg.DestType())
	}
	destNumber := msg.DestArg(FieldPhoneNumber)
	if destNumber == cfg.Twilio.FromNumber {
		return nil, errors.New("refusing to send outgoing SMS to FromNumber")
	}

	ctx = log.WithFields(ctx, log.Fields{
		"Phone": destNumber,
		"Type":  "TwilioSMS",
	})

	makeSMSCode := func(alertID int, serviceID string) int {
		if !hasTwoWaySMSSupport(ctx, destNumber) {
			return 0
		}

		code, err := s.b.insertDB(ctx, destNumber, msg.MsgID(), alertID, serviceID)
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "insert alert id for SMS callback -- sending 1-way SMS as fallback"))
			return 0
		}

		return code
	}

	var message string
	var err error
	switch t := msg.(type) {
	case notification.AlertStatus:
		message, err = renderAlertStatusMessage(cfg.ApplicationName(), t)
	case notification.AlertBundle:
		var link string
		if canContainURL(ctx, destNumber) {
			link = cfg.CallbackURL(fmt.Sprintf("/services/%s/alerts", t.ServiceID))
		}

		message, err = renderAlertBundleMessage(cfg.ApplicationName(), t, link, makeSMSCode(0, t.ServiceID))
	case notification.Alert:
		var link string
		if canContainURL(ctx, destNumber) {
			link = cfg.CallbackURL(fmt.Sprintf("/alerts/%d", t.AlertID))
		}

		message, err = renderAlertMessage(cfg.ApplicationName(), t, link, makeSMSCode(t.AlertID, ""))
	case notification.Test:
		message = fmt.Sprintf("%s: Test message.", cfg.ApplicationName())
	case notification.Verification:
		message = fmt.Sprintf("%s: Verification code: %s", cfg.ApplicationName(), t.Code)
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
	opts.CallbackParams.Set(msgParamID, msg.MsgID())
	// Actually send notification to end user & receive Message Status
	resp, err := s.c.SendSMS(ctx, destNumber, message, opts)
	if err != nil {
		return nil, errors.Wrap(err, "send message")
	}

	// If the message was sent successfully, reset reply limits.
	s.limit.Reset(destNumber)

	return resp.sentMessage(), nil
}

func (s *SMS) ServeStatusCallback(w http.ResponseWriter, req *http.Request) {
	if disabled(w, req) {
		return
	}
	ctx := req.Context()
	cfg := config.FromContext(ctx)
	status := MessageStatus(req.FormValue("MessageStatus"))
	sid := validSID(req.FormValue("MessageSid"))
	var number string
	if cfg.Twilio.RCSSenderID != "" && req.FormValue("To") == "rcs:"+cfg.Twilio.RCSSenderID {
		number = req.FormValue("To")
	} else {
		number = validPhone(req.FormValue("To"))
	}
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
	msg := Message{SID: sid, Status: status, From: strings.TrimPrefix(req.FormValue("From"), "rcs:")}

	log.Debugf(ctx, "Got Twilio SMS status callback.")

	err := s.r.SetMessageStatus(ctx, sid, msg.messageStatus())
	if err != nil {
		// log and continue
		log.Log(ctx, err)
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
	from := validPhone(strings.TrimPrefix(req.FormValue("From"), "rcs:"))
	if from == "" || from == cfg.Twilio.FromNumber || from == cfg.Twilio.RCSSenderID {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx = log.WithFields(ctx, log.Fields{
		"Number": from,
		"Type":   "TwilioSMS",
	})

	respond := func(isPassive bool, msg string) {
		if !isPassive {
			// always reset if an action was taken
			s.limit.Reset(from)
		}

		if s.limit.ShouldDrop(from) {
			log.Debugf(ctx, "SMS passive reply limit reached for %s, not replying.", from)
			return
		}

		if isPassive {
			valid, err := s.r.IsKnownDest(ctx, gadb.NewDestV1(DestTypeTwilioSMS, FieldPhoneNumber, from))
			if err != nil {
				log.Log(ctx, fmt.Errorf("check if known SMS number: %w", err))
			} else if !valid {
				// don't respond if the number is not known
				return
			}
			s.limit.RecordPassiveReply(from)
		}
		smsFrom := req.FormValue("To")
		if cfg.Twilio.MessagingServiceSID != "" {
			smsFrom = cfg.Twilio.MessagingServiceSID
		}
		_, err := s.c.SendSMS(ctx, from, msg, &SMSOptions{FromNumber: smsFrom})
		if err != nil {
			log.Log(ctx, errors.Wrap(err, "send response"))
		}
	}
	var err error
	retryOpts := []retry.Option{
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Second),
	}

	// handle start and stop codes from user
	body := req.FormValue("Body")
	if isStartMessage(body) {
		err := retry.DoTemporaryError(func(int) error { return s.r.Start(ctx, NewSMSDest(from)) }, retryOpts...)
		if err != nil {
			log.Log(ctx, fmt.Errorf("process START message: %w", err))
		}
		return
	}
	if isStopMessage(body) {
		err := retry.DoTemporaryError(func(int) error { return s.r.Stop(ctx, NewSMSDest(from)) }, retryOpts...)
		if err != nil {
			log.Log(ctx, fmt.Errorf("process STOP message: %w", err))
		}
		return
	}

	if cfg.Twilio.DisableTwoWaySMS {
		respond(true, "Response codes are currently disabled. Visit the dashboard to manage alerts.")
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
		} else if strings.HasPrefix(m[1], "e") {
			result = notification.ResultEscalate
		} else {
			result = notification.ResultResolve
		}
		lookupFn = func() (*codeInfo, error) { return s.b.LookupByCode(ctx, from, 0) }
	} else if m := shortReplyRx.FindStringSubmatch(body); len(m) == 3 {
		if strings.HasPrefix(m[2], "a") {
			result = notification.ResultAcknowledge
		} else if strings.HasPrefix(m[2], "e") {
			result = notification.ResultEscalate
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
		} else if strings.HasPrefix(m[1], "e") {
			result = notification.ResultEscalate
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
		} else if strings.HasPrefix(m[2], "e") {
			result = notification.ResultEscalate
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
		respond(true, "Sorry, but that isn't a request GoAlert understood. Visit the Web UI for more information. To unsubscribe, reply with STOP.")
		ctx = log.WithField(ctx, "SMSBody", body)
		log.Debug(ctx, errors.Wrap(err, "parse alert action"))
		return
	}

	var prefix string
	if result == notification.ResultAcknowledge {
		prefix = "Acknowledged"
	} else if result == notification.ResultEscalate {
		prefix = "Escalation requested"
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
		respond(true, "Unknown reply code for this action. Visit the dashboard to manage alerts.")
		return
	}

	msg := "System error. Visit the dashboard to manage alerts."
	if alert.IsAlreadyClosed(err) {
		nonSystemErr = true
		msg = fmt.Sprintf("Alert #%d already closed", alert.AlertID(err))
	} else if alert.IsAlreadyAcknowledged(err) {
		nonSystemErr = true
		msg = fmt.Sprintf("Alert #%d already acknowledged", alert.AlertID(err))
	} else if validation.IsClientError(err) {
		respond(true, "Error: "+stderrors.Unwrap(err).Error())
		return
	}

	if nonSystemErr {
		var e alert.LogEntryFetcher
		// alert store returns the special error struct, twilio checks if it's special, and if so, pulls the log entry
		if errors.As(err, &e) {
			// we pass a 'sudo' context to give permission
			permission.SudoContext(ctx, func(sCtx context.Context) {
				entry, err := e.LogEntry(sCtx)
				if err != nil {
					log.Log(sCtx, errors.Wrap(err, "fetch log entry"))
				} else {
					msg += "\n\n" + entry.String(ctx)
				}
			})
		} else {
			log.Log(ctx, errors.Wrap(err, "process notification response"))
		}
		respond(true, msg)
		return
	}

	if err != nil {
		log.Log(ctx, err)
		respond(true, msg)
		return
	}

	if info.ServiceName != "" {
		respond(false, fmt.Sprintf("%s all alerts for service '%s'", prefix, info.ServiceName))
	} else {
		respond(false, fmt.Sprintf("%s alert #%d", prefix, info.AlertID))
	}
}
