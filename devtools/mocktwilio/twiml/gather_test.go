package twiml_test

import (
	"encoding/xml"
	"os"

	"github.com/target/goalert/devtools/mocktwilio/twiml"
)

func ExampleGather() {
	var resp twiml.Response
	resp.Verbs = append(resp.Verbs, &twiml.Gather{
		Action:         "/gather",
		NumDigitsCount: 5,
		Verbs: []twiml.GatherVerb{
			&twiml.Say{
				Content: "Please enter your 5-digit zip code.",
			},
		},
	})

	_ = xml.NewEncoder(os.Stdout).Encode(resp)
	// Output:
	// <Response><Gather action="/gather" numDigits="5"><Say>Please enter your 5-digit zip code.</Say></Gather></Response>
}
