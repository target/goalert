package twiml

import (
	"encoding/xml"
	"strconv"
)

type NullInt struct {
	Valid bool
	Value int
}

var (
	_ xml.MarshalerAttr   = (*NullInt)(nil)
	_ xml.UnmarshalerAttr = (*NullInt)(nil)
)

func (n *NullInt) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Value == "" {
		n.Valid = false
		n.Value = 0
		return nil
	}

	val, err := strconv.Atoi(attr.Value)
	if err != nil {
		return err
	}

	n.Valid = true
	n.Value = val
	return nil
}

func (n *NullInt) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if n == nil || !n.Valid {
		return xml.Attr{}, nil
	}

	return xml.Attr{Name: name, Value: strconv.Itoa(n.Value)}, nil
}
