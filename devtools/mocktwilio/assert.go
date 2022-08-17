package mocktwilio

import (
	"context"
	"testing"
	"time"
)

type AssertConfig struct {
	ServerAPI
	Timeout        time.Duration
	AppPhoneNumber string

	// RefreshFunc will be called before waiting for new messages or calls to arrive.
	//
	// It is useful for testing purposes to ensure pending messages/calls are sent from the application.
	//
	// Implementations should not return until requests to mocktwilio are complete.
	RefreshFunc func()
}

type ServerAPI interface {
	SendMessage(ctx context.Context, from, to, body string) (Message, error)
	WaitInFlight(context.Context) error
	Messages() <-chan Message
	Calls() <-chan Call
}

func NewAssertions(t *testing.T, cfg AssertConfig) PhoneAssertions {
	return &assert{
		t:          t,
		assertBase: &assertBase{AssertConfig: cfg},
	}
}

func (a *assert) WithT(t *testing.T) PhoneAssertions {
	return &assert{
		t:          t,
		assertBase: a.assertBase,
	}
}

type assert struct {
	t *testing.T
	*assertBase
}

type assertBase struct {
	AssertConfig

	messages []Message
	calls    []Call

	ignoreSMS   []assertIgnore
	ignoreCalls []assertIgnore
}

type assertIgnore struct {
	number   string
	keywords []string
}

type texter interface {
	To() string
	Text() string
}
type answerer interface {
	Answer(context.Context) error
}

func (a *assert) matchMessage(destNumber string, keywords []string, t texter) bool {
	a.t.Helper()
	if t.To() != destNumber {
		return false
	}

	if ans, ok := t.(answerer); ok {
		ctx, cancel := context.WithTimeout(context.Background(), a.Timeout)
		defer cancel()

		err := ans.Answer(ctx)
		if err != nil {
			a.t.Fatalf("mocktwilio: error answering call to %s: %v", t.To(), err)
		}
	}

	return containsAll(t.Text(), keywords)
}

func (a *assert) refresh() {
	if a.RefreshFunc == nil {
		return
	}

	a.RefreshFunc()
}

func (a *assert) Device(number string) PhoneDevice {
	return &assertDev{a, number}
}

func (a *assert) WaitAndAssert() {
	a.t.Helper()

	// flush any remaining application messages
	a.refresh()
	a.ServerAPI.WaitInFlight(context.Background())

drainMessages:
	for {
		select {
		case msg := <-a.Messages():
			a.messages = append(a.messages, msg)
		default:
			break drainMessages
		}
	}

drainCalls:
	for {
		select {
		case call := <-a.Calls():
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
		a.t.Errorf("mocktwilio: unexpected SMS to %s: %s", msg.To(), msg.Text())
	}

checkCalls:
	for _, call := range a.calls {
		for _, ignore := range a.ignoreCalls {
			if a.matchMessage(ignore.number, ignore.keywords, call) {
				continue checkCalls
			}
		}

		hasFailure = true
		a.t.Errorf("mocktwilio: unexpected call to %s: %s", call.To(), call.Text())
	}

	if hasFailure {
		a.t.FailNow()
	}
}
