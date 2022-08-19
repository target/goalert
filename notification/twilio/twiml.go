package twilio

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/target/goalert/config"
)

type twiMLResponse struct {
	say []string

	voiceName     string
	voiceLanguage string

	gatherURL        string
	redirectURL      string
	redirectPauseSec int
	hangup           bool

	hasOptions     bool
	expectResponse bool

	sent bool

	w http.ResponseWriter
}

func newTwiMLResponse(ctx context.Context, w http.ResponseWriter) *twiMLResponse {
	cfg := config.FromContext(ctx)
	return &twiMLResponse{
		voiceName:     cfg.Twilio.VoiceName,
		voiceLanguage: cfg.Twilio.VoiceLanguage,
		w:             w,
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

type menuOption int

const (
	optionUnknown menuOption = iota
	optionCancel
	optionConfirmStop
	optionAck
	optionClose
	optionAckAll
	optionCloseAll
	optionStop
	optionRepeat
)

func (t *twiMLResponse) AddOptions(options ...menuOption) {
	t.hasOptions = true
	for _, opt := range options {
		switch opt {
		case optionConfirmStop:
			t.expectResponse = true
			t.Sayf("To confirm unenrollment of this number, press %s.", digitConfirm)
		case optionCancel:
			t.expectResponse = true
			t.Sayf("To go back to the previous menu, press %s.", digitGoBack)
		case optionStop:
			t.Sayf("To disable voice notifications to this number, press %s.", digitStop)
		case optionRepeat:
			t.Sayf("To repeat this message, press %s.", sayRepeat)
		case optionAck:
			t.expectResponse = true
			t.Sayf("To acknowledge, press %s.", digitAck)
		case optionClose:
			t.expectResponse = true
			t.Sayf("To close, press %s.", digitClose)
		case optionAckAll:
			t.expectResponse = true
			t.Sayf("To acknowledge all, press %s.", digitAck)
		case optionCloseAll:
			t.expectResponse = true
			t.Sayf("To close all, press %s.", digitClose)
		default:
			panic("Unknown option")
		}
	}
}

func (t *twiMLResponse) Gather(url string) {
	t.gatherURL = url
	if !t.expectResponse {
		t.Say("If you are done, you may simply hang up.")
	}
	t.AddOptions(optionRepeat)
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
	t.Say("Goodbye.")
	t.sendResponse()
}

func (t *twiMLResponse) sendResponse() {
	if t.sent {
		panic("Response already sent")
	}
	t.sent = true

	if t.hasOptions && t.gatherURL == "" {
		// if we give the user options, we need to allow them to respond
		panic("Options without gather")
	}

	t.w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	io.WriteString(t.w, xml.Header)
	io.WriteString(t.w, "<Response>\n")
	if t.gatherURL != "" {
		io.WriteString(t.w, `<Gather numDigits="1" timeout="10" action="`)
		xml.EscapeText(t.w, []byte(t.gatherURL))
		io.WriteString(t.w, `">`+"\n")
	}

	// add optional voice and/or language options
	voiceAndLanuage := ""
	if t.voiceName != "" && t.voiceLanguage != "" {
		// language is required with a voice
		voiceAndLanuage = fmt.Sprintf(` voice="%s" language="%s"`, t.voiceName, t.voiceLanguage)
	} else if t.voiceLanguage != "" {
		// language can be set without a voice
		voiceAndLanuage = fmt.Sprintf(` language="%s"`, t.voiceLanguage)
	}
	for _, s := range t.say {
		io.WriteString(t.w, fmt.Sprintf(`<Say%s><prosody rate="slow">`, voiceAndLanuage))
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
