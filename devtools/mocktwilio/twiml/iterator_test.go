package twiml

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIterator(t *testing.T) {
	const doc = `<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Gather numDigits="1" timeout="10" action="http://127.0.0.1:39111/api/v2/twilio/call?msgBody=VGhpcyBpcyBHb0FsZXJ0IHdpdGggYW4gYWxlcnQgbm90aWZpY2F0aW9uLiB0ZXN0aW5nLg&amp;msgID=62b6e835-e086-4cc2-9030-b0b230fdd2b2&amp;msgSubjectID=1&amp;type=alert">
		<Say>
			<prosody rate="slow">This is GoAlert with an alert notification. testing.</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To acknowledge, press 4.</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To close, press 6.</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To disable voice notifications to this number, press 1.</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To repeat this message, press star.</prosody>
		</Say>
	</Gather>
</Response>`

	i := NewIterator()
	err := i.SetResponse([]byte(doc))
	require.NoError(t, err)

	assert.True(t, i.Next())
	assert.Equal(t, &Say{Content: "\n\t\t\t<prosody rate=\"slow\">This is GoAlert with an alert notification. testing.</prosody>\n\t\t"}, i.Verb())

	assert.True(t, i.Next())
	assert.Equal(t, &Say{Content: "\n\t\t\t<prosody rate=\"slow\">To acknowledge, press 4.</prosody>\n\t\t"}, i.Verb())

	assert.True(t, i.Next())
	assert.Equal(t, &Say{Content: "\n\t\t\t<prosody rate=\"slow\">To close, press 6.</prosody>\n\t\t"}, i.Verb())

	assert.True(t, i.Next())
	assert.Equal(t, &Say{Content: "\n\t\t\t<prosody rate=\"slow\">To disable voice notifications to this number, press 1.</prosody>\n\t\t"}, i.Verb())

	assert.True(t, i.Next())
	assert.Equal(t, &Say{Content: "\n\t\t\t<prosody rate=\"slow\">To repeat this message, press star.</prosody>\n\t\t"}, i.Verb())

	assert.True(t, i.Next())
	assert.Equal(t, &Gather{
		Action:                      "http://127.0.0.1:39111/api/v2/twilio/call?msgBody=VGhpcyBpcyBHb0FsZXJ0IHdpdGggYW4gYWxlcnQgbm90aWZpY2F0aW9uLiB0ZXN0aW5nLg&msgID=62b6e835-e086-4cc2-9030-b0b230fdd2b2&msgSubjectID=1&type=alert",
		FinishOnKey:                 "#",
		Input:                       "dtmf",
		Language:                    "en-US",
		Method:                      "POST",
		NumDigitsCount:              1,
		PartialResultCallbackMethod: "POST",
		TimeoutDur:                  10 * time.Second,
		SpeechTimeoutDur:            10 * time.Second,
		SpeechModel:                 "default",
	}, i.Verb())

	assert.False(t, i.Next())
}
