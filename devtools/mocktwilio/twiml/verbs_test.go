package twiml

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSay(t *testing.T) {
	checkExp := func(exp Say, doc, expDoc string) {
		t.Helper()
		var s Say
		err := xml.Unmarshal([]byte(doc), &s)
		require.NoError(t, err)
		assert.Equal(t, exp, s)

		data, err := xml.Marshal(s)
		require.NoError(t, err)
		assert.Equal(t, expDoc, string(data))
	}
	check := func(exp Say, doc string) {
		t.Helper()
		checkExp(exp, doc, doc)
	}

	check(Say{Content: "hi"}, `<Say>hi</Say>`)
	checkExp(Say{Content: "hi"}, `<Say loop="">hi</Say>`, `<Say>hi</Say>`)
	checkExp(Say{Content: "hi", LoopCount: 1000}, `<Say loop="0">hi</Say>`, `<Say loop="1000">hi</Say>`)
	check(Say{Content: "hi", LoopCount: 1}, `<Say loop="1">hi</Say>`)
	check(Say{Content: "hi", LoopCount: 1000}, `<Say loop="1000">hi</Say>`)
	check(Say{Content: "hi", Voice: "foo"}, `<Say voice="foo">hi</Say>`)
}

func TestPause(t *testing.T) {
	checkExp := func(exp Pause, doc, expDoc string) {
		t.Helper()
		var s Pause
		err := xml.Unmarshal([]byte(doc), &s)
		require.NoError(t, err)
		assert.Equal(t, exp, s)

		data, err := xml.Marshal(s)
		require.NoError(t, err)
		assert.Equal(t, expDoc, string(data))
	}
	check := func(exp Pause, doc string) {
		t.Helper()
		checkExp(exp, doc, doc)
	}

	checkExp(Pause{Dur: time.Second}, `<Pause></Pause>`, `<Pause length="1"></Pause>`)
	checkExp(Pause{Dur: time.Second}, `<Pause/>`, `<Pause length="1"></Pause>`)
	check(Pause{Dur: 2 * time.Second}, `<Pause length="2"></Pause>`)
	check(Pause{Dur: time.Second}, `<Pause length="1"></Pause>`)
	check(Pause{}, `<Pause length="0"></Pause>`)
}
