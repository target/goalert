package twiml

import (
	"encoding/xml"
	"strconv"
	"time"
)

type Gather struct {
	// Action defaults to the current request URL and can be relative or absolute.
	Action string `xml:"action,attr,omitempty"`

	// FinishOnKey defaults to "#".
	FinishOnKey string `xml:"finishOnKey,attr,omitempty"`

	Hints string `xml:"hints,attr,omitempty"`

	// Input defaults to dtmf.
	Input string `xml:"input,attr,omitempty"`

	// Language defaults to en-US.
	Language string `xml:"language,attr,omitempty"`

	// Method defaults to POST.
	Method string `xml:"method,attr,omitempty"`

	// NumDigitsCount is the normalized version of `numDigits`. A zero value indicates no limit and is the default.
	NumDigitsCount int `xml:"-"`

	// PartialResultCallback defaults to the current request URL and can be relative or absolute.
	//
	// https://www.twilio.com/docs/voice/twiml/gather#partialresultcallback
	PartialResultCallback string `xml:"partialResultCallback,attr"`

	PartialResultCallbackMethod string `xml:"partialResultCallbackMethod,attr"`

	// ProfanityFilter defaults to true.
	ProfanityFilter bool `xml:"profanityFilter,attr"`

	// TimeoutDur defaults to 5 seconds.
	TimeoutDur time.Duration `xml:"-"`

	// SpeechTimeoutDur is the speechTimeout attribute.
	//
	// It should be ignored if SpeechTimeoutAuto is true. Defaults to TimeoutDur.
	SpeechTimeoutDur time.Duration `xml:"-"`

	// SpeechTimeoutAuto will be ture if the `speechTimeout` value is set to true.
	SpeechTimeoutAuto bool `xml:"-"`

	SpeechModel string `xml:"speechModel,attr,omitempty"`

	Enhanced bool `xml:"enhanced,attr,omitempty"`

	ActionOnEmptyResult bool `xml:"actionOnEmptyResult,attr,omitempty"`
}

func defStr(s *string, defaultValue string) {
	if *s == "" {
		*s = defaultValue
	}
}

func (g *Gather) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type RawGather Gather
	var gg struct {
		RawGather
		NumDigits       NullInt `xml:"numDigits,attr"`
		SpeechTimeout   string  `xml:"speechTimeout,attr"`
		Timeout         NullInt `xml:"timeout,attr"`
		ProfanityFilter *bool   `xml:"profanityFilter,attr"`
	}

	if err := d.DecodeElement(&gg, &start); err != nil {
		return err
	}
	*g = Gather(gg.RawGather)
	defStr(&g.FinishOnKey, "#")
	defStr(&g.Method, "POST")
	defStr(&g.Input, "dtmf")
	defStr(&g.Language, "en-US")
	defStr(&g.PartialResultCallbackMethod, "POST")
	defStr(&g.SpeechModel, "default")
	if gg.NumDigits.Valid {
		g.NumDigitsCount = gg.NumDigits.Value
		if g.NumDigitsCount < 1 {
			g.NumDigitsCount = 1
		}
	}
	if gg.ProfanityFilter != nil {
		g.ProfanityFilter = *gg.ProfanityFilter
	} else {
		g.ProfanityFilter = true
	}
	if gg.Timeout.Valid {
		g.TimeoutDur = time.Duration(gg.Timeout.Value) * time.Second
	} else {
		g.TimeoutDur = 5 * time.Second
	}
	switch gg.SpeechTimeout {
	case "auto":
		g.SpeechTimeoutAuto = true
	case "":
		g.SpeechTimeoutDur = g.TimeoutDur
	default:
		v, err := strconv.Atoi(gg.SpeechTimeout)
		if err != nil {
			return err
		}
		g.SpeechTimeoutDur = time.Duration(v) * time.Second
	}
	return nil
}

func (g Gather) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type RawGather Gather
	var gg struct {
		RawGather
		NumDigits     int    `xml:"numDigits,attr,omitempty"`
		SpeechTimeout string `xml:"speechTimeout,attr"`
		Timeout       int    `xml:"timeout,attr"`
	}
	if g.NumDigitsCount > 1 {
		gg.NumDigits = g.NumDigitsCount
	}
	if g.SpeechTimeoutAuto {
		gg.SpeechTimeout = "auto"
	} else {
		gg.SpeechTimeout = strconv.Itoa(int(g.SpeechTimeoutDur / time.Second))
	}
	gg.Timeout = int(g.TimeoutDur / time.Second)
	gg.RawGather = RawGather(g)
	return e.EncodeElement(gg, start)
}

type Hangup struct{}

// The Pause verb waits silently for a specific number of seconds.
//
// If Pause is the first verb in a TwiML document, Twilio will wait the specified number of seconds before picking up the call.
// https://www.twilio.com/docs/voice/twiml/pause
type Pause struct {
	// Dur derives from the `length` attribute and indicates the length of the pause.
	//
	// https://www.twilio.com/docs/voice/twiml/pause#attributes-length
	Dur time.Duration `xml:"-"`
}

func (p *Pause) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type PauseRaw Pause
	var pp struct {
		PauseRaw
		Length NullInt `xml:"length,attr"`
	}
	if err := d.DecodeElement(&pp, &start); err != nil {
		return err
	}
	*p = Pause(pp.PauseRaw)
	if pp.Length.Valid {
		p.Dur = time.Duration(pp.Length.Value) * time.Second
	} else {
		p.Dur = time.Second
	}
	return nil
}

func (p Pause) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type PauseRaw Pause
	var pp struct {
		PauseRaw
		Length *int `xml:"length,attr"`
	}
	sec := int(p.Dur / time.Second)
	pp.Length = &sec
	pp.PauseRaw = PauseRaw(p)
	return e.EncodeElement(pp, start)
}

type Redirect struct {
	// Method defaults to POST.
	//
	// https://www.twilio.com/docs/voice/twiml/redirect#attributes-method
	Method string `xml:"method,attr,omitempty"`

	// URL is the URL to redirect to.
	URL string `xml:",chardata"`
}

func (r *Redirect) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.DecodeElement(&r, &start); err != nil {
		return err
	}
	if r.Method == "" {
		r.Method = "POST"
	}
	return nil
}

type Reject struct {
	// Reason defaults to 'rejected'.
	//
	// https://www.twilio.com/docs/voice/twiml/reject#attributes-reason
	Reason string `xml:"reason,attr,omitempty"`
}

type RejectReason int

const (
	RejectReasonRejected RejectReason = iota
	RejectReasonBusy
)

func (r *Reject) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if err := d.DecodeElement(&r, &start); err != nil {
		return err
	}
	if r.Reason == "" {
		r.Reason = "rejected"
	}

	return nil
}

type Say struct {
	Content  string `xml:",innerxml"`
	Language string `xml:"language,attr,omitempty"`
	Voice    string `xml:"voice,attr,omitempty"`

	// LoopCount is derived from the `loop` attribute, taking into account the default value
	// and the special value of 0.
	//
	// If LoopCount is below zero, the message is not played.
	//
	// https://www.twilio.com/docs/voice/twiml/say#attributes-loop
	LoopCount int `xml:"-"`
}

var (
	_ xml.Unmarshaler = (*Say)(nil)
	_ xml.Marshaler   = Say{}
)

func (s *Say) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type SayRaw Say
	var ss struct {
		SayRaw
		Loop NullInt `xml:"loop,attr"`
	}
	if err := d.DecodeElement(&ss, &start); err != nil {
		return err
	}
	*s = Say(ss.SayRaw)
	if ss.Loop.Valid {
		s.LoopCount = ss.Loop.Value
		if s.LoopCount == 0 {
			s.LoopCount = 1000
		}
	}
	return nil
}

func (s Say) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	type SayRaw Say
	var ss struct {
		SayRaw
		Content string `xml:",innerxml"`
		Loop    *int   `xml:"loop,attr"`
	}
	ss.Content = s.Content
	if s.LoopCount != 0 {
		ss.Loop = &s.LoopCount
	}
	ss.SayRaw = SayRaw(s)
	return e.EncodeElement(ss, start)
}
