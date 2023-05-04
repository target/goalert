package twiml

import (
	"encoding/xml"
	"time"
)

// Hangup is a verb that hangs up the call.
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
		Length *int `xml:"length,attr,omitempty"`
	}
	sec := int(p.Dur / time.Second)
	pp.Length = &sec
	pp.PauseRaw = PauseRaw(p)
	return e.EncodeElement(pp, start)
}

// The Redirect verb redirects the current call to a new TwiML document.
type Redirect struct {
	// Method defaults to POST.
	//
	// https://www.twilio.com/docs/voice/twiml/redirect#attributes-method
	Method string `xml:"method,attr,omitempty"`

	// URL is the URL to redirect to.
	URL string `xml:",chardata"`
}

func (r *Redirect) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type RedirectRaw Redirect
	var rr struct {
		RedirectRaw
	}
	if err := d.DecodeElement(&rr, &start); err != nil {
		return err
	}
	*r = Redirect(rr.RedirectRaw)
	if r.Method == "" {
		r.Method = "POST"
	}
	return nil
}

// The Reject verb rejects an incoming call.
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

// Say is a TwiML verb that plays back a message to the caller.
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
