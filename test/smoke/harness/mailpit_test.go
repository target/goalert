package harness

import (
	"net/smtp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMailpit(t *testing.T) {
	mp, err := newMailpit(5)
	require.NoError(t, err)
	t.Cleanup(func() { _ = mp.Close() })

	err = smtp.SendMail(mp.smtpAddr, nil, "example@example.com", []string{"foo@bar.com"}, []byte("Subject: Hello\nTo: foo@bar.com\n\nWorld!"))
	require.NoError(t, err, "expected to be able to send email")

	err = smtp.SendMail(mp.smtpAddr, nil, "example@example.com", []string{"bin@baz.com"}, []byte("Subject: There\nTo: bin@baz.com\n\nThen!"))
	require.NoError(t, err, "expected to be able to send email")

	found := assert.Eventually(t, func() bool {
		found, err := mp.ReadMessage("foo@bar.com", "World")
		require.NoError(t, err, "expected to be able to read email")
		return found
	}, 5*time.Second, 100*time.Millisecond, "expected to find email to foo@bar.com containing 'World'")
	if !found {
		msgs, err := mp.UnreadMessages()
		require.NoError(t, err)
		t.Fatalf("timeout waiting for email; Got:\n%v", msgs)
	}

	found, err = mp.ReadMessage("foo@bar.com", "World")
	require.NoError(t, err, "expected to be able to read email")
	assert.False(t, found, "expected message to have been marked as read")

	remaining, err := mp.UnreadMessages()
	require.NoError(t, err)
	assert.Len(t, remaining, 1, "expected one unread message")
}
