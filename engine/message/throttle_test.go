package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/engine/message"
)

func TestThrottle(t *testing.T) {
	n := time.Now()
	check := func(desc string, cfg message.ThrottleConfig, times []time.Time, nextMessage time.Time) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			t.Helper()
			i := 0
			cur := n
			for {
				th := message.NewThrottle(cfg, cur, true)
				for _, ts := range times {
					th.Record(message.Message{SentAt: ts})
				}
				if !th.InCooldown(message.Message{}) {
					break
				}
				i++
				if i > 100 {
					t.Fatal("never found next message after 100 minutes")
				}
				cur = cur.Add(time.Minute)
			}
			assert.Equal(t, nextMessage.String(), cur.String())
		})
	}

	check("simple",
		message.ThrottleRules{{Count: 1, Per: time.Minute}},
		[]time.Time{n},
		n.Add(time.Minute),
	)

	check("burst",
		message.ThrottleRules{{Count: 1, Per: time.Minute}, {Count: 3, Per: 15 * time.Minute}},
		[]time.Time{n.Add(-10 * time.Minute), n.Add(-12 * time.Minute), n},
		n.Add(3*time.Minute),
	)

	check("not_at_cap",
		message.ThrottleRules{{Count: 3, Per: 15 * time.Minute}},
		[]time.Time{n},
		n,
	)

	check("smooth_rate",
		message.ThrottleRules{{Count: 3, Per: 15 * time.Minute, Smooth: true}},
		[]time.Time{n},
		n.Add(5*time.Minute),
	)

	check("smooth_burst",
		message.ThrottleRules{{Count: 3, Per: 10 * time.Minute}, {Count: 5, Per: 30 * time.Minute, Smooth: true}},
		[]time.Time{n, n, n},
		n.Add(10*time.Minute),
	)

}
