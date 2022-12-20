package twassert

import (
	"context"
	"fmt"
	"time"

	"github.com/target/goalert/devtools/mocktwilio"
)

type call struct {
	*dev
	mocktwilio.Call
}

func (a *assertions) newCall(baseCall mocktwilio.Call) *call {
	dev := &dev{a, baseCall.To()}
	return dev.newCall(baseCall)
}

func (d *dev) newCall(baseCall mocktwilio.Call) *call {
	call := &call{
		dev:  d,
		Call: baseCall,
	}
	d.t.Logf("mocktwilio: incoming %s", call)
	return call
}

func (d *dev) ExpectVoice(keywords ...string) {
	d.t.Helper()
	d.ExpectCall().Answer().ExpectSay(keywords...).Hangup()
}

// String returns a string representation of the call for test output.
func (c *call) String() string {
	return fmt.Sprintf("call from %s to %s", c.From(), c.To())
}

// Answer is part of the RingingCall interface.
func (c *call) Answer() ExpectedCall {
	c.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), c.dev.Timeout)
	defer cancel()

	err := c.Call.Answer(ctx)
	if err != nil {
		c.t.Fatalf("mocktwilio: answer %s: %v", c, err)
	}

	return c
}

func (c *call) Reject() { c.RejectWith(mocktwilio.CallFailed) }

func (c *call) RejectWith(status mocktwilio.FinalCallStatus) {
	c.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), c.dev.Timeout)
	defer cancel()

	err := c.Call.Hangup(ctx, status)
	if err != nil {
		c.t.Fatalf("mocktwilio: hangup %s with '%s': %v", c, status, err)
	}
}

func (c *call) Hangup() {
	c.t.Helper()
	c.RejectWith(mocktwilio.CallCompleted)
}

func (c *call) Press(digits string) ExpectedCall {
	c.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), c.dev.Timeout)
	defer cancel()

	err := c.Call.Press(ctx, digits)
	if err != nil {
		c.t.Fatalf("mocktwilio: press '%s' on %s: %v", digits, c, err)
	}

	return c
}

func (c *call) IdleForever() ExpectedCall {
	c.t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), c.dev.Timeout)
	defer cancel()

	err := c.PressTimeout(ctx)
	if err != nil {
		c.t.Fatalf("mocktwilio: wait on %s: %v", c, err)
	}

	return c
}

func (c *call) ExpectSay(keywords ...string) ExpectedCall {
	c.t.Helper()

	if !containsAll(c.Text(), keywords) {
		c.t.Fatalf("mocktwilio: expected %s to say: %v, but got: %s", c, keywords, c.Text())
	}

	return c
}

func (d *dev) ExpectCall() RingingCall {
	d.t.Helper()

	for idx, call := range d.calls {
		if call.To() != d.number {
			continue
		}

		// Remove the call from the list of calls.
		d.calls = append(d.calls[:idx], d.calls[idx+1:]...)

		return call
	}

	d.refresh()

	t := time.NewTimer(d.Timeout)
	defer t.Stop()

	ref := time.NewTicker(time.Second)
	defer ref.Stop()

	for {
		select {
		case <-t.C:
			d.t.Fatalf("mocktwilio: timeout after %s waiting for a voice call to %s", d.Timeout, d.number)
		case <-ref.C:
			d.refresh()
		case baseCall := <-d.Calls():
			call := d.newCall(baseCall)
			if call.To() != d.number {
				d.calls = append(d.calls, call)
				continue
			}

			return call
		}
	}
}
