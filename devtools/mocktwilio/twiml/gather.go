package twiml

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

// Gather is a TwiML verb that gathers pressed digits from the caller.
type Gather struct {
	XMLName xml.Name `xml:"Gather"`

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
	PartialResultCallback string `xml:"partialResultCallback,attr,omitempty"`

	PartialResultCallbackMethod string `xml:"partialResultCallbackMethod,attr,omitempty"`

	// DisableProfanityFilter disables the profanity filter.
	DisableProfanityFilter bool `xml:"-"`

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

	Verbs []GatherVerb `xml:"-"`
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

		Content []anyVerb `xml:",any"`
	}

	if err := d.DecodeElement(&gg, &start); err != nil {
		return err
	}
	*g = Gather(gg.RawGather)
	for _, v := range gg.Content {
		switch t := v.verb.(type) {
		case *Say:
			g.Verbs = append(g.Verbs, t)
		case *Pause:
			g.Verbs = append(g.Verbs, t)
		default:
			return fmt.Errorf("unexpected verb in Gather: %T", t)
		}
	}

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
		g.DisableProfanityFilter = !*gg.ProfanityFilter
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
		NumDigits       int    `xml:"numDigits,attr,omitempty"`
		SpeechTimeout   string `xml:"speechTimeout,attr,omitempty"`
		Timeout         int    `xml:"timeout,attr,omitempty"`
		ProfanityFilter *bool  `xml:"profanityFilter,attr,omitempty"`
		Verbs           []any  `xml:",any"`
	}
	gg.RawGather = RawGather(g)
	for _, v := range g.Verbs {
		switch t := v.(type) {
		case *Say:
			gg.Verbs = append(gg.Verbs, anyVerb{verb: t})
		case *Pause:
			gg.Verbs = append(gg.Verbs, anyVerb{verb: t})
		default:
			return fmt.Errorf("unexpected verb in Gather: %T", v)
		}
	}
	if g.NumDigitsCount > 1 {
		gg.NumDigits = g.NumDigitsCount
	}
	if g.SpeechTimeoutAuto {
		gg.SpeechTimeout = "auto"
	} else if g.SpeechTimeoutDur != gg.TimeoutDur {
		gg.SpeechTimeout = strconv.Itoa(int(g.SpeechTimeoutDur / time.Second))
	}
	if gg.Input == "dtmf" {
		gg.Input = ""
	}
	if gg.FinishOnKey == "#" {
		gg.FinishOnKey = ""
	}
	if gg.Language == "en-US" {
		gg.Language = ""
	}
	if gg.PartialResultCallback == "" || gg.PartialResultCallbackMethod == "POST" {
		gg.PartialResultCallbackMethod = ""
	}
	if gg.Method == "POST" {
		gg.Method = ""
	}
	if g.DisableProfanityFilter {
		gg.ProfanityFilter = new(bool)
		*gg.ProfanityFilter = false
	}
	if gg.SpeechModel == "default" {
		gg.SpeechModel = ""
	}
	gg.Timeout = int(g.TimeoutDur / time.Second)
	return e.EncodeElement(gg, start)
}
