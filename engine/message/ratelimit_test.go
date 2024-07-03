package message_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfy"
	"github.com/target/goalert/notification/twilio"
)

// TestRateLimit checks known good message sequences are allowed by the rate limit config.
func TestRateLimit(t *testing.T) {
	validate := func(desc string, msgType notification.MessageType, destType nfy.DestType, _times ...time.Time) {
		t.Helper()
		t.Run(desc, func(t *testing.T) {
			for i := 2; i <= len(_times); i++ {
				times := _times[:i]
				last := times[len(times)-1]
				th := message.NewThrottle(message.PerCMThrottle, last, false)
				for _, tm := range times[:len(times)-1] {
					th.Record(message.Message{Type: msgType, SentAt: tm, Dest: nfy.Dest{Type: destType}})
				}
				assert.Falsef(t, th.InCooldown(message.Message{Type: msgType, Dest: nfy.Dest{Type: destType}}), "message #%d should not be in cooldown", i)
			}
		})
	}

	validate("alert-voice",
		notification.MessageTypeAlert, twilio.DestTypeVoice,

		// {Count: 3, Per: 15 * time.Minute},
		// {Count: 7, Per: time.Hour, Smooth: true},
		// {Count: 15, Per: 3 * time.Hour, Smooth: true},

		time.Date(2015, time.May, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 0, 1, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 0, 2, 0, 0, time.UTC), // 15 min limit; 13 minute gap max

		time.Date(2015, time.May, 1, 0, 15, 0, 0, time.UTC),  // 15 min limit expired for first message
		time.Date(2015, time.May, 1, 0, 26, 15, 0, time.UTC), // per hour rule active
		time.Date(2015, time.May, 1, 0, 37, 30, 0, time.UTC), // 11.25 min max gap for hour
		time.Date(2015, time.May, 1, 0, 48, 45, 0, time.UTC),

		time.Date(2015, time.May, 1, 1, 0, 0, 0, time.UTC), // start of three hour window
		time.Date(2015, time.May, 1, 1, 15, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 1, 30, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 1, 45, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 2, 0, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 2, 15, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 2, 30, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 2, 45, 0, 0, time.UTC),
		time.Date(2015, time.May, 1, 3, 0, 0, 0, time.UTC),

		time.Date(2015, time.May, 1, 3, 1, 0, 0, time.UTC), // out of window
	)

	validate("alert-voice-staggered",
		notification.MessageTypeAlert, twilio.DestTypeVoice,

		// {Count: 3, Per: 15 * time.Minute},
		// {Count: 7, Per: time.Hour, Smooth: true},
		// {Count: 15, Per: 3 * time.Hour, Smooth: true},

		time.Date(2021, 7, 9, 2, 18, 20, 0, time.UTC),
		time.Date(2021, 7, 9, 3, 34, 25, 0, time.UTC),
		time.Date(2021, 7, 9, 4, 6, 25, 0, time.UTC),
		time.Date(2021, 7, 9, 4, 7, 30, 0, time.UTC),
		time.Date(2021, 7, 9, 4, 8, 35, 0, time.UTC),
		time.Date(2021, 7, 9, 4, 21, 28, 0, time.UTC),
		time.Date(2021, 7, 9, 4, 34, 30, 0, time.UTC),
		time.Date(2021, 7, 9, 5, 27, 25, 0, time.UTC),
		time.Date(2021, 7, 9, 5, 28, 33, 0, time.UTC),
		time.Date(2021, 7, 9, 6, 54, 40, 0, time.UTC),
		time.Date(2021, 7, 9, 7, 7, 31, 0, time.UTC),
		time.Date(2021, 7, 9, 7, 16, 40, 0, time.UTC),
		time.Date(2021, 7, 9, 7, 21, 30, 0, time.UTC),
		time.Date(2021, 7, 9, 7, 34, 35, 0, time.UTC),
		time.Date(2021, 7, 9, 8, 19, 40, 0, time.UTC),
		time.Date(2021, 7, 9, 8, 27, 28, 0, time.UTC),
		time.Date(2021, 7, 9, 8, 37, 0, 0, time.UTC),

		time.Date(2021, 7, 9, 8, 37+1, 0, 0, time.UTC),
	)
}
