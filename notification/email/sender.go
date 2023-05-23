package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/matcornic/hermes/v2"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"gopkg.in/gomail.v2"
)

type Sender struct{}

func NewSender(ctx context.Context) *Sender {
	return &Sender{}
}

var _ notification.Sender = &Sender{}

// Send will send an for the provided message type.
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	cfg := config.FromContext(ctx)

	fromAddr, err := mail.ParseAddress(cfg.SMTP.From)
	if err != nil {
		return nil, err
	}
	toAddr, err := mail.ParseAddress(msg.Destination().Value)
	if err != nil {
		return nil, err
	}
	if fromAddr.Name == "" {
		fromAddr.Name = cfg.ApplicationName()
	}

	h := hermes.Hermes{
		Product: hermes.Product{
			Name: cfg.ApplicationName(),
			Link: cfg.General.PublicURL,
			Logo: cfg.CallbackURL("/static/goalert-alt-logo.png"),
		},
	}
	var e hermes.Email
	var subject string
	switch m := msg.(type) {
	case notification.Test:
		subject = "Test Message"
		e.Body.Title = "Test Message"
		e.Body.Intros = []string{"This is a test message."}
	case notification.Verification:
		subject = "Verification Message"
		e.Body.Title = "Verification Message"
		e.Body.Intros = []string{"This is your contact method verification code."}
		e.Body.Actions = []hermes.Action{{
			Instructions: "Click the REACTIVATE link on your profile page and enter the verification code.",
			InviteCode:   strconv.Itoa(m.Code),
		}}
	case notification.Alert:
		subject = fmt.Sprintf("Alert #%d: %s", m.AlertID, m.Summary)
		e.Body.Title = fmt.Sprintf("Alert #%d", m.AlertID)
		e.Body.Intros = []string{m.Summary, m.Details}
		e.Body.Actions = []hermes.Action{{
			Button: hermes.Button{
				Text: "Open Alert Details",
				Link: cfg.CallbackURL(fmt.Sprintf("/alerts/%d", m.AlertID)),
			},
		}}
	case notification.AlertBundle:
		subject = fmt.Sprintf("Service %s has %d unacknowledged alerts", m.ServiceName, m.Count)
		e.Body.Title = "Multiple Unacknowledged Alerts"
		e.Body.Intros = []string{fmt.Sprintf("The service %s has %d unacknowledged alerts.", m.ServiceName, m.Count)}
		e.Body.Actions = []hermes.Action{{
			Button: hermes.Button{
				Text: "Open Alert List",
				Link: cfg.CallbackURL(fmt.Sprintf("/services/%s/alerts", m.ServiceID)),
			},
		}}
	case notification.AlertStatus:
		subject = fmt.Sprintf("Alert #%d: %s", m.AlertID, m.LogEntry)
		e.Body.Title = fmt.Sprintf("Alert #%d", m.AlertID)
		e.Body.Intros = []string{m.LogEntry}
		e.Body.Actions = []hermes.Action{{
			Button: hermes.Button{
				Text: "Open Alert Details",
				Link: cfg.CallbackURL(fmt.Sprintf("/alerts/%d", m.AlertID)),
			},
		}}
		e.Body.Outros = []string{"You are receiving this message because you have status updates enabled. Visit your Profile page to change this."}
	default:
		return nil, errors.New("message type not supported")
	}

	htmlBody, err := h.GenerateHTML(e)
	if err != nil {
		return nil, err
	}
	textBody, err := h.GeneratePlainText(e)
	if err != nil {
		return nil, err
	}

	g := gomail.NewMessage()
	g.SetHeader("From", fromAddr.String())
	g.SetAddressHeader("To", toAddr.Address, toAddr.Name)
	g.SetHeader("Subject", subject)
	g.SetBody("text/plain", textBody)
	g.AddAlternative("text/html", htmlBody)

	var buf bytes.Buffer

	_, err = g.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	host, port, _ := net.SplitHostPort(cfg.SMTP.Address)
	if host == "" {
		host = cfg.SMTP.Address
	}
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.SMTP.SkipVerify,
		ServerName:         host,
	}
	sendFn := SendMailTLS
	if cfg.SMTP.DisableTLS {
		sendFn = SendMail
		if port == "" {
			port = "25"
		}
	} else if port == "" {
		port = "465"
	}

	var authFn NegotiateAuth
	if cfg.SMTP.Username+cfg.SMTP.Password != "" {
		authFn = func(auths string) smtp.Auth {
			if strings.Contains(auths, "CRAM-MD5") {
				return smtp.CRAMMD5Auth(cfg.SMTP.Username, cfg.SMTP.Password)
			}
			if strings.Contains(auths, "PLAIN") {
				return smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, host)
			}
			if strings.Contains(auths, "LOGIN") {
				return LoginAuth(cfg.SMTP.Username, cfg.SMTP.Password, host)
			}

			return nil
		}
	}

	err = sendFn(ctx, net.JoinHostPort(host, port), authFn, fromAddr.Address, []string{toAddr.Address}, buf.Bytes(), tlsCfg)
	if err != nil {
		return nil, err
	}

	return &notification.SentMessage{
		State:    notification.StateSent,
		SrcValue: fromAddr.String(),
	}, nil
}
