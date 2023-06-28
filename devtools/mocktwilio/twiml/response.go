package twiml

import (
	"encoding/xml"
	"errors"
	"io"
)

type Response struct {
	Verbs []Verb `xml:"-"`
}

func (r *Response) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		t, err := d.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}
		if t == nil {
			break
		}

		if t, ok := t.(xml.StartElement); ok {
			v, err := decodeVerb(d, t)
			if err != nil {
				return err
			}
			r.Verbs = append(r.Verbs, v)
		}
	}

	return nil
}

func (r Response) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	for _, v := range r.Verbs {
		if err := encodeVerb(e, v); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}
