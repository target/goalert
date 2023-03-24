package twiml

import (
	"encoding/xml"
	"time"
)

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
