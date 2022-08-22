package twilio

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type twiMLResponse struct {
	say []string

	gatherURL        string
	redirectURL      string
	redirectPauseSec int
	hangup           bool

	hasOptions     bool
	expectResponse bool

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

type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Text    string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}
type twiMLDoc struct {
	XMLName xml.Name `xml:"Response"`
	xmlNode `xml:",any"`
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

	var doc twiMLDoc
	for _, s := range t.say {
		doc.Nodes = append(doc.Nodes, xmlNode{
			XMLName: xml.Name{Local: "Say"},
			Nodes: []xmlNode{
				{
					XMLName: xml.Name{Local: "prosody"},
					Attrs:   []xml.Attr{{Name: xml.Name{Local: "rate"}, Value: "slow"}},
					Text:    s,
				},
			},
		})
	}

	if t.redirectPauseSec > 0 {
		doc.Nodes = append(doc.Nodes, xmlNode{
			XMLName: xml.Name{Local: "Pause"},
			Attrs:   []xml.Attr{{Name: xml.Name{Local: "length"}, Value: strconv.Itoa(t.redirectPauseSec)}},
		})
	}

	if t.redirectURL != "" {
		doc.Nodes = append(doc.Nodes, xmlNode{
			XMLName: xml.Name{Local: "Redirect"},
			Text:    t.redirectURL,
		})
	}

	if t.gatherURL != "" {
		// wrap everything in a Gather node
		doc.Nodes = []xmlNode{
			{
				XMLName: xml.Name{Local: "Gather"},
				Attrs: []xml.Attr{
					{Name: xml.Name{Local: "numDigits"}, Value: "1"},
					{Name: xml.Name{Local: "timeout"}, Value: "10"},
					{Name: xml.Name{Local: "action"}, Value: t.gatherURL},
				},
				Nodes: doc.Nodes,
			},
		}
	}

	if t.hangup {
		doc.Nodes = append(doc.Nodes, xmlNode{
			XMLName: xml.Name{Local: "Hangup"},
		})
	}

	var buf bytes.Buffer
	io.WriteString(&buf, xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "\t")
	err := enc.Encode(doc)
	if err != nil {
		http.Error(t.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		panic(err)
	}

	t.w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	io.WriteString(t.w, buf.String())
}
