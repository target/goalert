package email

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/smtp"
	"strings"

	"github.com/target/goalert/util/log"
)

// validateLine checks to see if a line has CR or LF as per RFC 5321
func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return errors.New("smtp: A line must not contain CR or LF")
	}
	return nil
}

func validateAddrs(from string, to []string) error {
	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	return nil
}

// NegotiateAuth should return the appropriate smtp.Auth for the given server auth string.
type NegotiateAuth func(auths string) smtp.Auth

// SendMailTLS will send a message using the provided server over a TLS connection and optional auth.
func SendMailTLS(ctx context.Context, addr string, a NegotiateAuth, from string, to []string, msg []byte, cfg *tls.Config) error {
	err := validateAddrs(from, to)
	if err != nil {
		return err
	}
	t, _ := ctx.Deadline()
	conn, err := tls.DialWithDialer(&net.Dialer{Deadline: t}, "tcp", addr, cfg)
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(t)
	defer conn.Close()

	host, _, _ := net.SplitHostPort(addr)
	return sendMail(ctx, conn, host, a, from, to, msg, nil)
}

// SendMail will send a message using the provided server and optional auth. It will attempt to use STARTTLS if available from the server.
func SendMail(ctx context.Context, addr string, a NegotiateAuth, from string, to []string, msg []byte, cfg *tls.Config) error {
	err := validateAddrs(from, to)
	if err != nil {
		return err
	}

	t, _ := ctx.Deadline()
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(t)
	defer conn.Close()

	host, _, _ := net.SplitHostPort(addr)
	return sendMail(ctx, conn, host, a, from, to, msg, cfg)
}

func sendMail(ctx context.Context, conn net.Conn, host string, a NegotiateAuth, from string, to []string, msg []byte, cfg *tls.Config) error {
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer log.Close(ctx, c.Close)

	_, isTLS := conn.(*tls.Conn)
	if ok, _ := c.Extension("STARTTLS"); !isTLS && ok {
		if err = c.StartTLS(cfg); err != nil {
			return err
		}
	}
	if a != nil {
		ok, auths := c.Extension("AUTH")
		if !ok {
			return errors.New("notification/email: server doesn't support AUTH")
		}
		auth := a(auths)
		if auth == nil {
			return errors.New("notification/email: no supported AUTH mechanism available")
		}
		if err = c.Auth(auth); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return c.Quit()
}
