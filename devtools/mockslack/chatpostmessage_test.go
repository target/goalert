package mockslack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachmentsText(t *testing.T) {
	const data = `[{"color":"#862421","blocks":[{"type":"section","text":{"type":"mrkdwn","text":"\u003chttp://127.0.0.1:39999/alerts/1|Alert #1: testing\u003e"}},{"type":"section","text":{"type":"mrkdwn","text":"\u003e "}},{"type":"context","elements":[{"type":"plain_text","text":"Unacknowledged"}]}]}]`

	text, color, err := attachmentsText(data)
	assert.NoError(t, err)
	assert.Equal(t, "#862421", color)
	assert.Equal(t, "<http://127.0.0.1:39999/alerts/1|Alert #1: testing>\n> \nUnacknowledged\n", text)

}
