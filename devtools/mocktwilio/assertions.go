package mocktwilio

import (
	"context"
	"testing"
	"time"
)

type AssertConfig struct {
	ServerAPI
	// Timeout is used to set the timeout for all operations, expected messages/calls as well as API calls for things like answering a call.
	Timeout time.Duration

	// AppPhoneNumber is the phone number that the application will use to make calls and send messages.
	AppPhoneNumber string

	// RefreshFunc will be called before waiting for new messages or calls to arrive.
	//
	// It is useful for testing purposes to ensure pending messages/calls are sent from the application.
	//
	// Implementations should not return until requests to mocktwilio are complete.
	RefreshFunc func()
}

// ServerAPI is the interface for the mocktwilio server.
type ServerAPI interface {
	SendMessage(ctx context.Context, from, to, body string) (Message, error)

	// WaitInFlight should return after all in-flight messages are processed.
	WaitInFlight(context.Context) error

	Messages() <-chan Message
	Calls() <-chan Call
}

func NewAssertions(t *testing.T, cfg AssertConfig) PhoneAssertions {
	return &assertions{
		t:          t,
		assertBase: &assertBase{AssertConfig: cfg},
	}
}

func (a *assertions) WithT(t *testing.T) PhoneAssertions {
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
	AssertConfig

	messages []*assertSMS
	calls    []*assertCall

	ignoreSMS []assertIgnore
}

type assertIgnore struct {
	number   string
	keywords []string
}

func (a *assertions) matchMessage(destNumber string, keywords []string, t *assertSMS) bool {
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

// Device will allow expecting calls and messages from a particular destination number.
func (a *assertions) Device(number string) PhoneDevice {
	return &assertDev{a, number}
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
			call := a.newAssertCall(baseCall)
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
