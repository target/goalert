package mocktwilio

import (
	"context"
	"time"
)

type assertCall struct {
	*assertDev
	Call
}

func (dev *assertDev) ExpectVoice(keywords ...string) {
	dev.t.Helper()
	dev.ExpectCall().Answer().ExpectSay(keywords...).Hangup()
}

func (call *assertCall) Answer() ExpectedCall {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()

	err := call.Call.Answer(ctx)
	if err != nil {
		call.t.Fatalf("mocktwilio: error answering call to %s: %v", call.To(), err)
	}

	return call
}

func (call *assertCall) Reject() {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()

	err := call.Call.Hangup(ctx, CallFailed)
	if err != nil {
		call.t.Fatalf("mocktwilio: error answering call to %s: %v", call.To(), err)
	}
}

func (call *assertCall) Hangup() {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()

	err := call.Call.Hangup(ctx, CallCompleted)
	if err != nil {
		call.t.Fatalf("mocktwilio: error ending call to %s: %v", call.To(), err)
	}
}

func (call *assertCall) Press(digits string) ExpectedCall {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()
	err := call.Call.Press(ctx, digits)
	if err != nil {
		call.t.Fatalf("mocktwilio: error pressing digits %s to %s: %v", digits, call.To(), err)
	}

	return call
}

func (call *assertCall) IdleForever() ExpectedCall {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.assertDev.Timeout)
	defer cancel()
	err := call.PressTimeout(ctx)
	if err != nil {
		call.t.Fatalf("mocktwilio: error waiting to %s: %v", call.To(), err)
	}

	return call
}

func (call *assertCall) ExpectSay(keywords ...string) ExpectedCall {
	call.t.Helper()

	if !containsAll(call.Text(), keywords) {
		call.t.Fatalf("mocktwilio: expected call to %s to contain keywords: %v, but got: %s", call.To(), keywords, call.Text())
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

		return &assertCall{
			assertDev: dev,
			Call:      call,
		}
	}

	dev.refresh()

	t := time.NewTimer(dev.Timeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			dev.t.Fatalf("mocktwilio: timeout after %s waiting for a voice call to %s", dev.Timeout, dev.number)
		case call := <-dev.Calls():
			dev.t.Logf("mocktwilio: incoming call from %s to %s", call.From(), call.To())
			if call.To() != dev.number {
				dev.calls = append(dev.calls, call)
				continue
			}

			return &assertCall{
				assertDev: dev,
				Call:      call,
			}
		}
	}
}
