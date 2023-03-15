package twilio

import (
	"bytes"
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
	optionEscalate
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
		case optionEscalate:
			t.expectResponse = true
			t.Sayf("To escalate, press %s.", digitEscalate)
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

type verbSay struct {
	XMLName  xml.Name `xml:"Say"`
	Language string   `xml:"language,attr,omitempty"`
	Voice    string   `xml:"voice,attr,omitempty"`
	Text     string   `xml:"-"`
	Prosody  *prosody `xml:"prosody"`
}
type prosody struct {
	XMLName xml.Name `xml:"prosody"`
	Text    string   `xml:",chardata"`
	Rate    string   `xml:"rate,attr"`
}

type twimlResponse struct {
	XMLName xml.Name `xml:"Response"`
	Verbs   []any    `xml:",any"`
}
type verbPause struct {
	XMLName   xml.Name `xml:"Pause"`
	LengthSec int      `xml:"length,attr"`
}
type verbRedirect struct {
	XMLName xml.Name `xml:"Redirect"`
	URL     string   `xml:",chardata"`
}
type verbHangup struct {
	XMLName xml.Name `xml:"Hangup"`
}
type verbGather struct {
	XMLName    xml.Name `xml:"Gather"`
	NumDigits  int      `xml:"numDigits,attr"`
	TimeoutSec int      `xml:"timeout,attr"`
	Action     string   `xml:"action,attr"`
	Verbs      []any    `xml:",any"`
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

	var doc twimlResponse
	for _, text := range t.say {
		doc.Verbs = append(
			doc.Verbs,
			verbSay{
				Language: t.voiceLanguage,
				Voice:    t.voiceName,
				Prosody: &prosody{
					Rate: "slow",
					Text: text,
				},
			},
		)
	}

	if t.redirectPauseSec > 0 {
		doc.Verbs = append(doc.Verbs, verbPause{LengthSec: t.redirectPauseSec})
	}

	if t.redirectURL != "" {
		doc.Verbs = append(doc.Verbs, verbRedirect{URL: t.redirectURL})
	}

	if t.gatherURL != "" {
		doc.Verbs = []any{verbGather{
			Action:     t.gatherURL,
			TimeoutSec: 10,
			NumDigits:  1,
			Verbs:      doc.Verbs,
		}}
	}

	if t.hangup {
		doc.Verbs = append(doc.Verbs, verbHangup{})
	}

	var buf bytes.Buffer
	_, _ = io.WriteString(&buf, xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "\t")
	err := enc.Encode(doc)
	if err != nil {
		http.Error(t.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		panic(err)
	}

	t.w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = io.WriteString(t.w, buf.String())
}
