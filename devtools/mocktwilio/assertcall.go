package mocktwilio

import (
	"context"
	"fmt"
	"time"
)

type assertCall struct {
	*assertDev
	Call
}

func (a *assertions) newAssertCall(baseCall Call) *assertCall {
	dev := &assertDev{a, baseCall.From()}
	return dev.newAssertCall(baseCall)
}

func (dev *assertDev) newAssertCall(baseCall Call) *assertCall {
	call := &assertCall{
		assertDev: dev,
		Call:      baseCall,
	}
	dev.t.Logf("mocktwilio: incoming %s", call)
	return call
}

func (dev *assertDev) ExpectVoice(keywords ...string) {
	dev.t.Helper()
	dev.ExpectCall().Answer().ExpectSay(keywords...).Hangup()
}

// String returns a string representation of the call for test output.
func (call *assertCall) String() string {
	return fmt.Sprintf("call from %s to %s", call.From(), call.To())
}

// Answer is part of the RingingCall interface.
func (call *assertCall) Answer() ExpectedCall {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()

	err := call.Call.Answer(ctx)
	if err != nil {
		call.t.Fatalf("mocktwilio: answer %s: %v", call, err)
	}

	return call
}

func (call *assertCall) Reject() { call.RejectWith(CallFailed) }

func (call *assertCall) RejectWith(status FinalCallStatus) {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()

	err := call.Call.Hangup(ctx, status)
	if err != nil {
		call.t.Fatalf("mocktwilio: hangup %s with '%s': %v", call, status, err)
	}
}

func (call *assertCall) Hangup() {
	call.t.Helper()
	call.RejectWith(CallCompleted)
}

func (call *assertCall) Press(digits string) ExpectedCall {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()

	err := call.Call.Press(ctx, digits)
	if err != nil {
		call.t.Fatalf("mocktwilio: press '%s' on %s: %v", digits, call, err)
	}

	return call
}

func (call *assertCall) IdleForever() ExpectedCall {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()

	err := call.PressTimeout(ctx)
	if err != nil {
		call.t.Fatalf("mocktwilio: wait on %s: %v", call, err)
	}

	return call
}

func (call *assertCall) ExpectSay(keywords ...string) ExpectedCall {
	call.t.Helper()

	if !containsAll(call.Text(), keywords) {
		call.t.Fatalf("mocktwilio: expected %s to say: %v, but got: %s", call, keywords, call.Text())
	}

	return call
}

func (dev *assertDev) ExpectCall() RingingCall {
	dev.t.Helper()

	for idx, call := range dev.calls {
		if call.To() != dev.number {
			continue
		}

		// Remove the call from the list of calls.
		dev.calls = append(dev.calls[:idx], dev.calls[idx+1:]...)

		return call
	}

	dev.refresh()

	t := time.NewTimer(dev.Timeout)
	defer t.Stop()

	ref := time.NewTicker(time.Second)
	defer ref.Stop()

	for {
		select {
		case <-t.C:
			dev.t.Fatalf("mocktwilio: timeout after %s waiting for a voice call to %s", dev.Timeout, dev.number)
		case <-ref.C:
			dev.refresh()
		case baseCall := <-dev.Calls():
			call := dev.newAssertCall(baseCall)
			if call.To() != dev.number {
				dev.calls = append(dev.calls, call)
				continue
			}

			return call
		}
	}
}
