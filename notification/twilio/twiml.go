package twilio

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type twiMLResponse struct {
	say []string

	gatherURL        string
	redirectURL      string
	redirectPauseSec int
	hangup           bool

	sent bool

	w http.ResponseWriter
}

func newTwiMLResponse(w http.ResponseWriter) *twiMLResponse {
	return &twiMLResponse{
		w: w,
	}
}

func (t *twiMLResponse) Redirect(url string) {
	t.redirectURL = url
	t.sendResponse()
}
func (t *twiMLResponse) RedirectPauseSec(url string, seconds int) {
	t.redirectURL = url
	t.redirectPauseSec = seconds
	t.sendResponse()
}

func (t *twiMLResponse) Gather(url string) {
	t.gatherURL = url
	t.sendResponse()
}

func (t *twiMLResponse) SayUnknownDigit() *twiMLResponse {
	t.Say("Sorry, I didn't understand that.")
	return t
}

func (t *twiMLResponse) Say(text string) *twiMLResponse {
	t.say = append(t.say, text)
	return t
}
func (t *twiMLResponse) Sayf(format string, args ...interface{}) *twiMLResponse {
	return t.Say(fmt.Sprintf(format, args...))
}

func (t *twiMLResponse) Hangup() {
	t.hangup = true
	t.sendResponse()
}

func (t *twiMLResponse) sendResponse() {
	if t.sent {
		panic("Response already sent")
	}
	t.sent = true

	// always offer repeat on gather
	if t.gatherURL != "" {
		t.Sayf("To repeat this message, press %s.", sayRepeat)
	}

	t.w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	io.WriteString(t.w, xml.Header)
	io.WriteString(t.w, "<Response>\n")
	if t.gatherURL != "" {
		io.WriteString(t.w, `<Gather numDigits="1" timeout="10" action="`)
		xml.EscapeText(t.w, []byte(t.gatherURL))
		io.WriteString(t.w, `">`+"\n")
	}
	for _, s := range t.say {
		io.WriteString(t.w, `<Say><prosody rate="slow">`)
		xml.EscapeText(t.w, []byte(s))
		io.WriteString(t.w, "</prosody></Say>\n")
	}

	if t.redirectPauseSec > 0 {
		fmt.Fprintf(t.w, `<Pause length="%d"/>`+"\n", t.redirectPauseSec)
	}

	if t.redirectURL != "" {
		io.WriteString(t.w, "<Redirect>")
		xml.EscapeText(t.w, []byte(t.redirectURL))
		io.WriteString(t.w, "</Redirect>\n")
	}
	if t.gatherURL != "" {
		io.WriteString(t.w, "</Gather>\n")
	}
	if t.hangup {
		io.WriteString(t.w, "<Hangup/>\n")
	}
	io.WriteString(t.w, "</Response>\n")
}
