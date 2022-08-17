package mocktwilio

import (
	"context"
	"time"
)

type assertCall struct {
	*assertDev
	Call
}

func (call *assertCall) Hangup() {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.Timeout)
	defer cancel()

	err := call.End(ctx, CallCompleted)
	if err != nil {
		call.t.Fatalf("mocktwilio: error ending call to %s: %v", call.To(), err)
	}
}

func (call *assertCall) ThenPress(digits string) ExpectedCall {
	call.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), call.Timeout)
	defer cancel()
	err := call.Press(ctx, digits)
	if err != nil {
		call.t.Fatalf("mocktwilio: error pressing digits %s to %s: %v", digits, call.To(), err)
	}

	return call
}

func (call *assertCall) ThenExpect(keywords ...string) ExpectedCall {
	call.t.Helper()

	if !containsAll(call.Text(), keywords) {
		call.t.Fatalf("mocktwilio: expected call to %s to contain keywords: %v, but got: %s", call.To(), keywords, call.Text())
	}

	return call
}

func (dev *assertDev) ExpectVoice(keywords ...string) ExpectedCall {
	dev.t.Helper()

	return &assertCall{
		assertDev: dev,
		Call:      dev.getVoice(false, keywords),
	}
}

func (dev *assertDev) RejectVoice(keywords ...string) {
	dev.t.Helper()

	call := dev.getVoice(false, keywords)
	ctx, cancel := context.WithTimeout(context.Background(), dev.Timeout)
	defer cancel()

	err := call.End(ctx, CallFailed)
	if err != nil {
		dev.t.Fatalf("mocktwilio: error ending call to %s: %v", call.To(), err)
	}
}

func (dev *assertDev) IgnoreUnexpectedVoice(keywords ...string) {
	dev.ignoreCalls = append(dev.ignoreCalls, assertIgnore{number: dev.number, keywords: keywords})
}

func (dev *assertDev) getVoice(prev bool, keywords []string) Call {
	dev.t.Helper()

	if prev {
		for _, call := range dev.calls {
			if !dev.matchMessage(dev.number, keywords, call) {
				continue
			}

			return call
		}
	}

	dev.refresh()

	t := time.NewTimer(dev.Timeout)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			dev.t.Fatalf("mocktwilio: timeout after %s waiting for a voice call to %s with keywords: %v", dev.Timeout, dev.number, keywords)
		case call := <-dev.Calls():
			if !dev.matchMessage(dev.number, keywords, call) {
				dev.calls = append(dev.calls, call)
				continue
			}

			return call
		}
	}
}
