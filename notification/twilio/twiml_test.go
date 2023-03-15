package twilio

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/config"
)

func TestTwiMLResponse(t *testing.T) {
	t.Run("hangup", func(t *testing.T) {
		var mockConfig config.Config
		ctx := mockConfig.Context(context.Background())
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(ctx, rec)
		r.Say("Hello")
		r.Hangup()

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Say>
		<prosody rate="slow">Hello</prosody>
	</Say>
	<Say>
		<prosody rate="slow">Goodbye.</prosody>
	</Say>
	<Hangup></Hangup>
</Response>`, string(data))
	})

	t.Run("redirect", func(t *testing.T) {
		var mockConfig config.Config
		mockConfig.Twilio.VoiceLanguage = "en-US"
		ctx := mockConfig.Context(context.Background())
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(ctx, rec)
		r.Say("Hello")
		r.Redirect("http://example.com")

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Say language="en-US">
		<prosody rate="slow">Hello</prosody>
	</Say>
	<Redirect>http://example.com</Redirect>
</Response>`, string(data))
	})

	t.Run("redirect-pause", func(t *testing.T) {
		var mockConfig config.Config
		mockConfig.Twilio.VoiceName = "Polly.Joanna-Neural"
		mockConfig.Twilio.VoiceLanguage = "en-US"
		ctx := mockConfig.Context(context.Background())
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(ctx, rec)
		r.Say("Hello! This is GoAlert.")
		r.RedirectPauseSec("http://example.com", 3)

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Say language="en-US" voice="Polly.Joanna-Neural">
		<prosody rate="slow">Hello! This is GoAlert.</prosody>
	</Say>
	<Pause length="3"></Pause>
	<Redirect>http://example.com</Redirect>
</Response>`, string(data))
	})

	t.Run("unknown-gather", func(t *testing.T) {
		var mockConfig config.Config
		ctx := mockConfig.Context(context.Background())
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(ctx, rec)
		r.SayUnknownDigit()
		r.Say("Hello")
		r.Gather("http://example.com")

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Gather numDigits="1" timeout="10" action="http://example.com">
		<Say>
			<prosody rate="slow">Sorry, I didn&#39;t understand that.</prosody>
		</Say>
		<Say>
			<prosody rate="slow">Hello</prosody>
		</Say>
		<Say>
			<prosody rate="slow">If you are done, you may simply hang up.</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To repeat this message, press star.</prosody>
		</Say>
	</Gather>
</Response>`, string(data))
	})

	t.Run("ack test", func(t *testing.T) {
		var mockConfig config.Config
		ctx := mockConfig.Context(context.Background())
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(ctx, rec)
		r.Say("Hello")
		r.AddOptions(optionAck)
		r.Gather("http://example.com")

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Gather numDigits="1" timeout="10" action="http://example.com">
		<Say>
			<prosody rate="slow">Hello</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To acknowledge, press 4.</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To repeat this message, press star.</prosody>
		</Say>
	</Gather>
</Response>`, string(data))
	})
	t.Run("esc test", func(t *testing.T) {
		var mockConfig config.Config
		ctx := mockConfig.Context(context.Background())
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(ctx, rec)
		r.Say("Hello")
		r.AddOptions(optionEscalate)
		r.Gather("http://example.com")

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Gather numDigits="1" timeout="10" action="http://example.com">
		<Say>
			<prosody rate="slow">Hello</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To escalate, press 5.</prosody>
		</Say>
		<Say>
			<prosody rate="slow">To repeat this message, press star.</prosody>
		</Say>
	</Gather>
</Response>`, string(data))
	})
}
