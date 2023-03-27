package twiml

import (
	"encoding/xml"
	"errors"
	"fmt"
)

type (
	v          struct{}
	GatherVerb interface {
		Verb
		isGatherVerb() v
	}
	Verb interface {
		isVerb() v
	}
)
func (*Say) isGatherVerb() v   { return v{} }

func (*Say) isVerb() v      { return v{} }
func (*Gather) isVerb() v   { return v{} }
type anyVerb struct {
	verb Verb
}

func decodeVerb(d *xml.Decoder, start xml.StartElement) (Verb, error) {
	switch start.Name.Local {
	case "Gather":
		g := new(Gather)
		return g, d.DecodeElement(&g, &start)
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
