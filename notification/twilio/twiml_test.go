package twilio

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTwiMLResponse(t *testing.T) {
	t.Run("hangup", func(t *testing.T) {
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(rec)
		r.Say("Hello")
		r.Hangup()

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
<Say><prosody rate="slow">Hello</prosody></Say>
<Hangup/>
</Response>
`, string(data))
	})

	t.Run("redirect", func(t *testing.T) {
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(rec)
		r.Say("Hello")
		r.Redirect("http://example.com")

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
<Say><prosody rate="slow">Hello</prosody></Say>
<Redirect>http://example.com</Redirect>
</Response>
`, string(data))
	})

	t.Run("redirect-pause", func(t *testing.T) {
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(rec)
		r.Say("Hello")
		r.RedirectPauseSec("http://example.com", 3)

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
<Say><prosody rate="slow">Hello</prosody></Say>
<Pause length="3"/>
<Redirect>http://example.com</Redirect>
</Response>
`, string(data))
	})

	t.Run("unknown-gather", func(t *testing.T) {
		rec := httptest.NewRecorder()

		r := newTwiMLResponse(rec)
		r.SayUnknownDigit()
		r.Say("Hello")
		r.Gather("http://example.com")

		resp := rec.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Type"), "application/xml")
		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<Response>
<Gather numDigits="1" timeout="10" action="http://example.com">
<Say><prosody rate="slow">Sorry, I didn&#39;t understand that.</prosody></Say>
<Say><prosody rate="slow">Hello</prosody></Say>
<Say><prosody rate="slow">To repeat this message, press star.</prosody></Say>
</Gather>
</Response>
`, string(data))
	})

}
