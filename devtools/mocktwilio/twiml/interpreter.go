package twiml

import (
	"encoding/xml"
	"fmt"
)

type Interpreter struct {
	state Verb

	verbs []Verb
}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (i *Interpreter) SetResponse(data []byte) error {
	var rsp Response
	if err := xml.Unmarshal(data, &rsp); err != nil {
		return err
	}

	i.verbs = append(i.verbs[:0], rsp.Verbs...)
	i.state = nil
	return nil
}

func (i *Interpreter) Verb() Verb { return i.state }

func (i *Interpreter) Next() bool {
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

		newVerbs := make([]any, 0, len(t.Verbs)+len(i.verbs))
		i.state = t.Verbs[0]
		for _, v := range t.Verbs[1:] {
			newVerbs = append(newVerbs, v)
		}
		t.Verbs = nil
		newVerbs = append(newVerbs, t)
		for _, v := range i.verbs[1:] {
			newVerbs = append(newVerbs, v)
		}
	case *Pause, *Say:
		i.state = t
		i.verbs = i.verbs[1:]
	default:
		panic(fmt.Sprintf("unhandled verb: %T", t))
	}

	return true
}
