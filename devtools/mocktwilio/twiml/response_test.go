package twiml

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponse_UnmarshalXML(t *testing.T) {
	const doc = `<Response><Say>hi</Say><Gather timeout="13"><Say>there</Say></Gather><Hangup></Hangup></Response>`

	var r Response
	err := xml.Unmarshal([]byte(doc), &r)
	require.NoError(t, err)

	assert.Equal(t, Response{Verbs: []Verb{
		&Say{Content: "hi"},
		&Gather{
			TimeoutDur: 13 * time.Second,
			Verbs:      []GatherVerb{&Say{Content: "there"}},

			// defaults should be set
			FinishOnKey:                 "#",
			Input:                       "dtmf",
			Language:                    "en-US",
			Method:                      "POST",
			PartialResultCallbackMethod: "POST",
			SpeechTimeoutDur:            13 * time.Second, // defaults to TimeoutDur
			SpeechModel:                 "default",
		},
		&Hangup{},
	}}, r)

	data, err := xml.Marshal(r)
	require.NoError(t, err)
	assert.Equal(t, doc, string(data))
}
