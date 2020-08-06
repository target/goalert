package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/mail"
	"strconv"

	"gopkg.in/gomail.v2"

	"github.com/matcornic/hermes/v2"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
)

type Sender struct{}

func NewSender(ctx context.Context) *Sender {
	return &Sender{}
}

var _ notification.Sender = &Sender{}

// Send will send an for the provided message type.
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {
	cfg := config.FromContext(ctx)

	fromAddr, err := mail.ParseAddress(cfg.SMTP.From)
	if err != nil {
		return nil, err
	}
	toAddr, err := mail.ParseAddress(msg.Destination().Value)
	if err != nil {
		return nil, err
	}

	h := hermes.Hermes{
		Product: hermes.Product{
			Name: "GoAlert",
			Link: cfg.General.PublicURL,
			Logo: "",
		},
	}
	var e hermes.Email
	var subject string
	switch m := msg.(type) {
	case notification.Test:
		e.Body.Title = "Test Message"
		e.Body.Intros = []string{"This is a test message from GoAlert."}
	case notification.Verification:
		e.Body.Title = "Verification Message"
		e.Body.Intros = []string{"This is a verification message from GoAlert."}
		e.Body.Actions = []hermes.Action{{
			Instructions: "Click the REACTIVATE link on your profile page and enter the verification code.",
			InviteCode:   strconv.Itoa(m.Code),
			Button: hermes.Button{
				Link: cfg.CallbackURL("/profile"),
			},
		}}
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

	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.SMTP.SkipVerify,
	}

	host, port, _ := net.SplitHostPort(cfg.SMTP.Address)
	sendFn := SendMailTLS
	if cfg.SMTP.DisableTLS {
		sendFn = SendMail
		if port == "" {
			port = "25"
		}
	} else if port == "" {
		port = "587"
	}

	var authFn NegotiateAuth

	err = sendFn(ctx, net.JoinHostPort(host, port), authFn, fromAddr.Address, []string{toAddr.Address}, buf.Bytes(), tlsCfg)
	if err != nil {
		return nil, err
	}

	return &notification.MessageStatus{ID: msg.ID(), State: notification.MessageStateSent}, nil
}

// Status is not supported by the email provider and will aways return an error.
func (s *Sender) Status(ctx context.Context, id, providerID string) (*notification.MessageStatus, error) {
	return nil, errors.New("notification/email: status not supported")
}
