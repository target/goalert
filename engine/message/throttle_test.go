package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/engine/message"
)

func TestThrottle(t *testing.T) {
	n := time.Now()
	check := func(desc string, cfg message.ThrottleConfig, times []time.Time, nextMessages ...time.Time) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			sentTimes := make([]time.Time, len(times))
			copy(sentTimes, times)
			t.Helper()
			i := 0
			cur := n
			for _, nextMessage := range nextMessages {
				for {
					th := message.NewThrottle(cfg, cur, true)
					for _, ts := range sentTimes {
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

				assert.Equal(t, nextMessage.In(n.Location()).Format(time.UnixDate), cur.Format(time.UnixDate))
				sentTimes = append(sentTimes, nextMessage)
			}
		})
	}

	check("simple",
		message.ThrottleRules{{Count: 1, Per: time.Minute}},
		[]time.Time{n},
		n.Add(time.Minute),
		n.Add(2*time.Minute),
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
		n.Add(10*time.Minute),
	)

	check("smooth_burst",
		message.ThrottleRules{{Count: 3, Per: 15 * time.Minute}, {Count: 18, Per: 60 * time.Minute, Smooth: true}},
		[]time.Time{n, n, n},
		n.Add(15*time.Minute),
		n.Add(18*time.Minute),
	)

	t.Run("max delay issue", func(t *testing.T) {
		cfg := message.ThrottleRules{
			{Count: 3, Per: 15 * time.Minute},                 // max delay = 12 minutes
			{Count: 7, Per: 60 * time.Minute, Smooth: true},   // max delay = (7-3=4) per (60-15=45) min; 45/4 = 11.25 min delay
			{Count: 15, Per: 180 * time.Minute, Smooth: true}, //max delay = (15-7=8) per (180-60=120) min; 120/8 = 15 min max delay
		}

		times := []time.Time{
			time.Date(2021, 7, 9, 8, 37, 0, 0, time.UTC),
			time.Date(2021, 7, 9, 8, 27, 28, 0, time.UTC),
			time.Date(2021, 7, 9, 8, 19, 40, 0, time.UTC),
			time.Date(2021, 7, 9, 7, 34, 35, 0, time.UTC),
			time.Date(2021, 7, 9, 7, 21, 30, 0, time.UTC),
			time.Date(2021, 7, 9, 7, 16, 40, 0, time.UTC),
			time.Date(2021, 7, 9, 7, 7, 31, 0, time.UTC),
			time.Date(2021, 7, 9, 6, 54, 40, 0, time.UTC),
			time.Date(2021, 7, 9, 5, 28, 33, 0, time.UTC),
			time.Date(2021, 7, 9, 5, 27, 25, 0, time.UTC),
			time.Date(2021, 7, 9, 4, 34, 30, 0, time.UTC),
			time.Date(2021, 7, 9, 4, 21, 28, 0, time.UTC),
			time.Date(2021, 7, 9, 4, 8, 35, 0, time.UTC),
		}

		th := message.NewThrottle(cfg, time.Date(2021, 7, 9, 8, 57, 0, 0, time.UTC), true)
		for _, ts := range times {
			th.Record(message.Message{SentAt: ts})
		}
		isThrottled := th.InCooldown(message.Message{})
		assert.False(t, isThrottled, "should not be throttled")
	})

}
