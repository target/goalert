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
	shifts := []oncall.Shift{{
		UserID: "01020304-0506-0708-090a-0b0c0d0e0f10",
		Start:  time.Date(2020, 1, 1, 8, 0, 0, 0, time.UTC),
		End:    time.Date(2020, 1, 15, 8, 0, 0, 0, time.UTC),
	}}
	generatedAt := time.Date(2020, 1, 1, 5, 0, 0, 0, time.UTC)
	cs.ScheduleID = "100f0e0d-0c0b-0a09-0807-060504030201"
	iCal, err := cs.renderICalFromShifts(shifts, generatedAt)
	assert.NoError(t, err)
	expected := strings.Join([]string{
		"BEGIN:VCALENDAR",
		"PRODID:-//GoAlert//dev//EN",
		"VERSION:2.0",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		"UID:4c7d37bf28d64eccc1e74a3889cfc97f6839a00fa781c91721058df915de27ce",
		"SUMMARY:On-Call Shift",
		"DTSTAMP:20200101T050000Z",
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
