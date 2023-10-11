package twiml

import (
	"encoding/xml"
	"fmt"
)

// Iterator will iterate over the verbs in a TwiML document,
// in the order they should be processed.
//
// A Say verb within a Gather verb will be processed before
// the Gather verb itself.
type Iterator struct {
	state Verb

	verbs []Verb
}

// NewIterator returns a new Iterator.
func NewIterator() *Iterator {
	return &Iterator{}
}

// SetResponse sets the response to interpret, and resets the state.
//
// This method will return an error if the response is not valid TwiML.
func (i *Iterator) SetResponse(data []byte) error {
	var rsp Response
	if err := xml.Unmarshal(data, &rsp); err != nil {
		return err
	}

	i.verbs = append(i.verbs[:0], rsp.Verbs...)
	i.state = nil
	return nil
}

// Verb returns the current verb.
func (i *Iterator) Verb() Verb { return i.state }

// Next advances the iterator to the next verb
// and returns true if there is another verb.
func (i *Iterator) Next() bool {
	if len(i.verbs) == 0 {
		return false
	}

	switch t := i.verbs[0].(type) {
	case *Hangup, *Redirect, *Reject:
		// end the call
		i.state = t
		i.verbs = nil
	case *Gather:
		if len(t.Verbs) == 0 {
			i.state = t
			i.verbs = i.verbs[1:]
			break
		}

		newVerbs := make([]Verb, 0, len(t.Verbs)+len(i.verbs))
		i.state = t.Verbs[0]
		for _, v := range t.Verbs[1:] {
			newVerbs = append(newVerbs, v)
		}
		t.Verbs = nil
		newVerbs = append(newVerbs, t)
		newVerbs = append(newVerbs, i.verbs[1:]...)
		i.verbs = newVerbs
	case *Pause, *Say:
		i.state = t
		i.verbs = i.verbs[1:]
	default:
		panic(fmt.Sprintf("unhandled verb: %T", t))
	}

	return true
}
