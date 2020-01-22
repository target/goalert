package calendarsubscription

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/oncall"
)

func TestCalendarSubscription_RenderICalFromShifts(t *testing.T) {
	var cs CalendarSubscription
	cs.Config.ReminderMinutes = []int{5, 10}
	shifts := []oncall.Shift{{Start: time.Date(2020, 1, 1, 8, 0, 0, 0, time.UTC), End: time.Date(2020, 1, 15, 8, 0, 0, 0, time.UTC)}}

	iCal, err := cs.renderICalFromShifts(shifts)
	assert.NoError(t, err)
	expected := strings.Join([]string{
		"BEGIN:VCALENDAR",
		"PRODID:-//GoAlert//dev//EN",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		"SUMMARY:On-Call Shift",
		"DTSTART:20200101T080000Z",
		"DTEND:20200115T080000Z",
		"BEGIN:VALARM",
		"ACTION:DISPLAY",
		"DESCRIPTION:REMINDER",
		"TRIGGER:-PT5M",
		"END:VALARM",
		"BEGIN:VALARM",
		"ACTION:DISPLAY",
		"DESCRIPTION:REMINDER",
		"TRIGGER:-PT10M",
		"END:VALARM",
		"END:VEVENT",
		"END:VCALENDAR",
		"",
	}, "\r\n")
	assert.Equal(t, expected, string(iCal))
}
