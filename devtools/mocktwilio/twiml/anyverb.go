package twiml

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type (
	v struct{}

	// GatherVerb is any verb that can be nested inside a Gather.
	GatherVerb interface {
		Verb
		isGatherVerb() v // unexported method to prevent external implementations
	}

	// Verb is any verb that can be nested inside a Response.
	Verb interface {
		isVerb() v // unexported method to prevent external implementations
	}
)

func (*Pause) isGatherVerb() v { return v{} }
func (*Say) isGatherVerb() v   { return v{} }

func (*Pause) isVerb() v    { return v{} }
func (*Say) isVerb() v      { return v{} }
func (*Gather) isVerb() v   { return v{} }
func (*Hangup) isVerb() v   { return v{} }
func (*Redirect) isVerb() v { return v{} }
func (*Reject) isVerb() v   { return v{} }

type anyVerb struct {
	verb Verb
}

func decodeVerb(d *xml.Decoder, start xml.StartElement) (Verb, error) {
	switch start.Name.Local {
	case "Gather":
		g := new(Gather)
		return g, d.DecodeElement(&g, &start)
	case "Hangup":
		h := new(Hangup)
		return h, d.DecodeElement(&h, &start)
	case "Pause":
		p := new(Pause)
		return p, d.DecodeElement(&p, &start)
	case "Redirect":
		r := new(Redirect)
		return r, d.DecodeElement(&r, &start)
	case "Reject":
		re := new(Reject)
		return re, d.DecodeElement(&re, &start)
	case "Say":
		s := new(Say)
		return s, d.DecodeElement(&s, &start)
	}

	return nil, errors.New("unsupported verb: " + start.Name.Local)
}

func encodeVerb(e *xml.Encoder, v Verb) error {
	var name string
	switch v.(type) {
	case *Gather:
		name = "Gather"
	case *Hangup:
		name = "Hangup"
	case *Pause:
		name = "Pause"
	case *Redirect:
		name = "Redirect"
	case *Reject:
		name = "Reject"
	case *Say:
		name = "Say"
	default:
		return fmt.Errorf("unsupported verb: %T", v)
	}

	if err := e.EncodeElement(v, xml.StartElement{Name: xml.Name{Local: name}}); err != nil {
		return err
	}

	return nil
}

func (v *anyVerb) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	v.verb, err = decodeVerb(d, start)
	return err
}

func (v anyVerb) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	return encodeVerb(e, v.verb)
}
