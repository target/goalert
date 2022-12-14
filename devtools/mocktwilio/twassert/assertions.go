package twassert

import (
	"context"
	"testing"
)

func NewAssertions(t *testing.T, cfg Config) Assertions {
	return &assertions{
		t:          t,
		assertBase: &assertBase{Config: cfg},
	}
}

func (a *assertions) WithT(t *testing.T) Assertions {
	return &assertions{
		t:          t,
		assertBase: a.assertBase,
	}
}

type assertions struct {
	t *testing.T
	*assertBase
}

type assertBase struct {
	Config

	messages []*sms
	calls    []*call

	ignoreSMS []ignoreRule
}

type ignoreRule struct {
	number   string
	keywords []string
}

func (a *assertions) matchMessage(destNumber string, keywords []string, t *sms) bool {
	a.t.Helper()
	if t.To() != destNumber {
		return false
	}

	return containsAll(t.Text(), keywords)
}

func (a *assertions) refresh() {
	if a.RefreshFunc == nil {
		return
	}

	a.RefreshFunc()
}

// WaitAndAssert will ensure no unexpected messages or calls are received.
func (a *assertions) WaitAndAssert() {
	a.t.Helper()

	// flush any remaining application messages
	a.refresh()
	a.ServerAPI.WaitInFlight(context.Background())

drainMessages:
	for {
		select {
		case msg := <-a.Messages():
			sms := a.newAssertSMS(msg)
			a.messages = append(a.messages, sms)
		default:
			break drainMessages
		}
	}

drainCalls:
	for {
		select {
		case baseCall := <-a.Calls():
			call := a.newCall(baseCall)
			a.calls = append(a.calls, call)
		default:
			break drainCalls
		}
	}

	var hasFailure bool

checkMessages:
	for _, msg := range a.messages {
		for _, ignore := range a.ignoreSMS {
			if a.matchMessage(ignore.number, ignore.keywords, msg) {
				continue checkMessages
			}
		}

		hasFailure = true
		a.t.Errorf("mocktwilio: unexpected %s", msg)
	}

	for _, call := range a.calls {
		hasFailure = true
		a.t.Errorf("mocktwilio: unexpected %s", call)
	}

	if hasFailure {
		a.t.FailNow()
	}
}
