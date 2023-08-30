package smtpsrv

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

func TestSession_Auth(t *testing.T) {
	var sess Session
	err := sess.AuthPlain("", "")
	assert.ErrorContains(t, err, "not supported")
}

func TestSession_Mail(t *testing.T) {
	var sess Session
	err := sess.Mail("", nil)
	assert.ErrorContains(t, err, "sender address")

	err = sess.Mail("test", nil)
	assert.ErrorContains(t, err, "sender address")

	err = sess.Mail("test@localhost", nil)
	assert.NoError(t, err)
}

func TestSession_Rcpt(t *testing.T) {
	var sess Session
	sess.cfg.MaxRecipients = 1
	sess.cfg.BackgroundContext = func() context.Context { return log.WithLogger(context.Background(), log.NewLogger()) }

	err := sess.Rcpt("")
	assert.ErrorContains(t, err, "recipient address")

	err = sess.Rcpt("test")
	assert.ErrorContains(t, err, "recipient address")

	err = sess.Rcpt("test@localhost")
	assert.ErrorContains(t, err, "domain not handled here")

	sess.cfg.Domain = "localhost"
	err = sess.Rcpt("test@localhost")
	assert.ErrorContains(t, err, "recipient address") // must be uuid

	authCtx := sess.cfg.BackgroundContext()

	sess.cfg.AuthorizeFunc = func(ctx context.Context, id string) (context.Context, error) {
		t.Helper()
		assert.Equal(t, "00000000-0000-0000-0000-000000000000", id)
		return authCtx, nil
	}

	err = sess.Rcpt("00000000-0000-0000-0000-000000000000+dedup-value@localhost")
	assert.NoError(t, err)
	assert.Contains(t, sess.authCtx, authCtx)
	assert.Equal(t, "dedup-value", sess.dedup)

	errFailed := errors.New("failed")
	sess.cfg.AuthorizeFunc = func(ctx context.Context, id string) (context.Context, error) {
		return sess.cfg.BackgroundContext(), errFailed
	}

	sess.authCtx = nil
	err = sess.Rcpt("00000000-0000-0000-0000-000000000000+dedup-value@localhost")
	assert.ErrorContains(t, err, "local error")
}

func TestSession_Data(t *testing.T) {
	var sess Session
	err := sess.Data(bytes.NewReader(nil))
	assert.ErrorContains(t, err, "recipient")

	sess.authCtx = []context.Context{permission.ServiceContext(context.Background(), "svc")}
	sess.from = "test@localhost"
	sess.dedup = "dedup-value"

	var createdAlert bool
	sess.cfg.CreateAlertFunc = func(ctx context.Context, a *alert.Alert) error {
		t.Helper()
		createdAlert = true
		assert.Equal(t, "dedup-value", a.DedupKey().Payload)
		assert.Equal(t, "test", a.Summary)
		assert.Equal(t, "From: test@localhost\n\nHello, world!", a.Details)
		assert.Equal(t, alert.SourceEmail, a.Source)
		assert.Equal(t, alert.StatusTriggered, a.Status)
		assert.Equal(t, permission.ServiceID(ctx), a.ServiceID)
		return nil
	}

	err = sess.Data(strings.NewReader("Subject: test\r\n\r\nHello, world!"))
	assert.NoError(t, err)
	assert.True(t, createdAlert, "CreateAlertFunc not called")
}
