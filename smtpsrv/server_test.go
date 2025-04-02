package smtpsrv_test

import (
	"context"
	"net"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/smtpsrv"
)

func TestServer(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })

	var lastAlert *alert.Alert
	srv := smtpsrv.NewServer(smtpsrv.Config{
		Domain:        "localhost",
		MaxRecipients: 1,
		AuthorizeFunc: func(ctx context.Context, id string) (context.Context, error) {
			return permission.ServiceContext(ctx, "svc"), nil
		},
		CreateAlertFunc: func(ctx context.Context, a *alert.Alert) error {
			lastAlert = a
			return nil
		},
		BackgroundContext: func() context.Context { return context.Background() },
	})

	go func() { _ = srv.ServeSMTP(l) }()
	t.Cleanup(func() { _ = srv.Shutdown(context.Background()) })

	c, err := smtp.Dial(l.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { c.Close() })

	err = smtp.SendMail(l.Addr().String(), nil, "test@localhost", []string{"test@localhost"}, []byte("test"))
	assert.ErrorContains(t, err, "recipient address")

	// test with uuid
	err = smtp.SendMail(l.Addr().String(), nil, "test@localhost", []string{"00000000-0000-0000-0000-000000000000+dedup-value@localhost"}, []byte("Subject: test\r\n\r\nbody"))
	assert.NoError(t, err)
	require.NotNil(t, lastAlert)
	assert.Equal(t, "test", lastAlert.Summary)
	assert.Equal(t, "From: <test@localhost>\n\nbody", lastAlert.Details)
	assert.Equal(t, "dedup-value", lastAlert.DedupKey().Payload)
	assert.Equal(t, alert.SourceEmail, lastAlert.Source)
	assert.Equal(t, alert.StatusTriggered, lastAlert.Status)
	assert.Equal(t, "svc", lastAlert.ServiceID)
}
