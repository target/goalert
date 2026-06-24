package twilio

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/util/log"
)

// DefaultTwilioAPIURL is the value that will be used for API calls if Config.BaseURL is empty.
const DefaultTwilioAPIURL = "https://api.twilio.com"

// SMSOptions allows configuring outgoing SMS messages.
type SMSOptions struct {
	// ValidityPeriod controls how long a message will still be valid in Twilio's queue.
	ValidityPeriod time.Duration

	// CallbackParams will be added to callback URLs
	CallbackParams url.Values

	// FromNumber allows overriding the specified FromNumber instead of using the context config.
	FromNumber string
}

// VoiceOptions allows configuring outgoing voice calls.
type VoiceOptions struct {
	// ValidityPeriod controls how long a message will still be valid in Twilio's queue.
	ValidityPeriod time.Duration

	// The call type/message.
	CallType CallType

	// CallbackParams will be added to all callback URLs
	CallbackParams url.Values

	// Params will be added to the voice callback URL
	Params url.Values
}

func (sms *SMSOptions) apply(v url.Values) {
	if sms == nil {
		return
	}
	if sms.ValidityPeriod != 0 {
		v.Set("ValidityPeriod", strconv.FormatFloat(sms.ValidityPeriod.Seconds(), 'f', -1, 64))
	}
}

func (voice *VoiceOptions) apply(v url.Values) {
	if voice == nil {
		return
	}
	if voice.ValidityPeriod != 0 {
		v.Set("ValidityPeriod", strconv.FormatFloat(voice.ValidityPeriod.Seconds(), 'f', -1, 64))
	}
}

func urlJoin(base string, parts ...string) string {
	base = strings.TrimSuffix(base, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(strings.Trim(p, "/"))
	}
	return base + "/" + strings.Join(parts, "/")
}

func (c *Config) url(parts ...string) string {
	base := c.BaseURL
	if base == "" {
		base = DefaultTwilioAPIURL
	}
	return urlJoin(urlJoin(base, "2010-04-01"), parts...)
}

func (c *Config) httpClient() *http.Client {
	if c.Client != nil {
		return c.Client
	}

	return http.DefaultClient
}

func (c *Config) get(ctx context.Context, urlStr string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	cfg := config.FromContext(ctx)
	req = req.WithContext(ctx)
	req.Header.Set("X-Twilio-Signature", string(Signature(cfg.Twilio.AuthToken, urlStr, nil)))
	req.SetBasicAuth(cfg.Twilio.AccountSID, cfg.Twilio.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return c.httpClient().Do(req)
}

func (c *Config) post(ctx context.Context, urlStr string, v url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", urlStr, bytes.NewBufferString(v.Encode()))
	if err != nil {
		return nil, err
	}
	cfg := config.FromContext(ctx)
	req = req.WithContext(ctx)
	req.Header.Set("X-Twilio-Signature", string(Signature(cfg.Twilio.AuthToken, urlStr, v)))
	req.SetBasicAuth(cfg.Twilio.AccountSID, cfg.Twilio.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return c.httpClient().Do(req)
}

// GetSMS will return the current state of a Message from Twilio.
func (c *Config) GetSMS(ctx context.Context, sid string) (*Message, error) {
	cfg := config.FromContext(ctx)
	urlStr := c.url("Accounts", cfg.Twilio.AccountSID, "Messages", sid+".json")
	resp, err := c.get(ctx, urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		var e Exception
		err = json.Unmarshal(data, &e)
		if err != nil {
			return nil, errors.Wrap(err, "parse error response")
		}
		return nil, &e
	}

	var m Message
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, errors.Wrap(err, "parse message response")
	}

	return &m, nil
}

// GetVoice will return the current state of a voice call from Twilio.
func (c *Config) GetVoice(ctx context.Context, sid string) (*Call, error) {
	cfg := config.FromContext(ctx)
	urlStr := c.url("Accounts", cfg.Twilio.AccountSID, "Calls", sid+".json")
	resp, err := c.post(ctx, urlStr, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		var e Exception
		err = json.Unmarshal(data, &e)
		if err != nil {
			return nil, errors.Wrap(err, "parse error response")
		}
		return nil, &e
	}

	var v Call
	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, errors.Wrap(err, "parse voice call response")
	}
	return &v, nil
}

// CallbackURL will return the callback url for the given configuration.
func (voice *VoiceOptions) CallbackURL(cfg config.Config) (string, error) {
	if voice == nil {
		voice = &VoiceOptions{}
	}
	if voice.CallType == "" {
		return "", errors.New("CallType missing")
	}
	return cfg.CallbackURL("/api/v2/twilio/call?type="+url.QueryEscape(string(voice.CallType)), voice.CallbackParams, voice.Params), nil
}

// setMsgBody will encode and set/update the message body parameter.
func (voice *VoiceOptions) setMsgBody(body string) {
	if voice.Params == nil {
		voice.Params = make(url.Values)
	}

	// Encode the body, so we don't need to worry about
	// buggy apps not escaping URL params properly.
	voice.Params.Set(msgParamBody, b64enc.EncodeToString([]byte(body)))
}

// setMsgParams will set parameters for the provided message.
func (voice *VoiceOptions) setMsgParams(msg notification.Message) (err error) {
	if voice.CallbackParams == nil {
		voice.CallbackParams = make(url.Values)
	}
	if voice.Params == nil {
		voice.Params = make(url.Values)
	}

	subID := -1
	switch t := msg.(type) {
	case notification.AlertBundle:
		voice.Params.Set(msgParamBundle, "1")
		voice.CallType = CallTypeAlert
	case notification.Alert:
		voice.CallType = CallTypeAlert
		subID = t.AlertID
	case notification.AlertStatus:
		voice.CallType = CallTypeAlertStatus
		subID = t.AlertID
	case notification.Test:
		voice.CallType = CallTypeTest
	case notification.Verification:
		voice.CallType = CallTypeVerify
	default:
		return errors.Errorf("unhandled message type: %T", t)
	}

	voice.Params.Set(msgParamSubID, strconv.Itoa(subID))
	voice.CallbackParams.Set(msgParamID, msg.MsgID())

	return nil
}

// StatusCallbackURL will return the status callback url for the given configuration.
func (voice *VoiceOptions) StatusCallbackURL(cfg config.Config) (string, error) {
	if voice == nil {
		voice = &VoiceOptions{}
	}
	return cfg.CallbackURL("/api/v2/twilio/call/status", voice.CallbackParams), nil
}

// StatusCallbackURL will return the status callback url for the given configuration.
func (sms *SMSOptions) StatusCallbackURL(cfg config.Config) (string, error) {
	if sms == nil {
		sms = &SMSOptions{}
	}
	return cfg.CallbackURL("/api/v2/twilio/message/status", sms.CallbackParams), nil
}

// StartVoice will initiate a voice call to the given number.
func (c *Config) StartVoice(ctx context.Context, to string, o *VoiceOptions) (*Call, error) {
	cfg := config.FromContext(ctx)
	v := make(url.Values)
	v.Set("To", to)
	v.Set("From", cfg.Twilio.FromNumber)
	stat, err := o.StatusCallbackURL(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "build status callback URL")
	}
	v.Set("StatusCallback", stat)

	voiceCallbackURL, err := o.CallbackURL(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "build voice callback URL")
	}
	v.Set("Url", voiceCallbackURL)
	v.Set("FallbackUrl", voiceCallbackURL)
	v.Add("StatusCallbackEvent", "initiated")
	v.Add("StatusCallbackEvent", "ringing")
	v.Add("StatusCallbackEvent", "answered")
	v.Add("StatusCallbackEvent", "completed")
	o.apply(v)
	urlStr := c.url("Accounts", cfg.Twilio.AccountSID, "Calls.json")

	resp, err := c.post(ctx, urlStr, v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 201 {
		var e Exception
		err = json.Unmarshal(data, &e)
		if err != nil {
			return nil, errors.Wrap(err, "parse error response")
		}
		return nil, &e
	}

	var call Call
	err = json.Unmarshal(data, &call)
	if err != nil {
		return nil, errors.Wrap(err, "parse voice call response")
	}
	if call.ErrorMessage != nil && call.ErrorCode != nil {
		return &call, &Exception{
			Status:  resp.StatusCode,
			Message: *call.ErrorMessage,
			Code:    int(*call.ErrorCode),
		}
	}
	return &call, nil
}

// SendSMS will send an SMS using Twilio.
func (c *Config) SendSMS(ctx context.Context, to, body string, o *SMSOptions) (*Message, error) {
	if o == nil {
		o = &SMSOptions{}
	}
	cfg := config.FromContext(ctx)
	v := make(url.Values)
	v.Set("To", to)
	if o.FromNumber != "" {
		v.Set("From", o.FromNumber)
	} else {
		info, err := c.CarrierInfo(ctx, to, cfg.Twilio.SMSCarrierLookup)
		if err != nil && cfg.Twilio.SMSCarrierLookup {
			log.Log(ctx, err)
		}
		if info != nil {
			v.Set("From", cfg.TwilioSMSFromNumber(info.Name))
		} else {
			v.Set("From", cfg.TwilioSMSFromNumber(""))
		}
	}
	v.Set("Body", body)

	stat, err := o.StatusCallbackURL(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "build status callback URL")
	}
	v.Set("StatusCallback", stat)
	o.apply(v)
	urlStr := c.url("Accounts", cfg.Twilio.AccountSID, "Messages.json")

	resp, err := c.post(ctx, urlStr, v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 201 {
		var e Exception
		err = json.Unmarshal(data, &e)
		if err != nil {
			return nil, errors.Wrap(err, "parse error response")
		}
		return nil, &e
	}

	var m Message
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, errors.Wrap(err, "parse message response")
	}
	if m.ErrorCode != nil && m.ErrorMessage != nil {
		return &m, &Exception{
			Status:  resp.StatusCode,
			Message: *m.ErrorMessage,
			Code:    int(*m.ErrorCode),
		}
	}

	return &m, nil
}
