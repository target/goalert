package calsub

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/oncall"
)

func TestRenderData_RenderICal(t *testing.T) {
	shifts := []oncall.Shift{{
		UserID: "01020304-0506-0708-090a-0b0c0d0e0f10",
		Start:  time.Date(2020, 1, 1, 8, 0, 0, 0, time.UTC),
		End:    time.Date(2020, 1, 15, 8, 0, 0, 0, time.UTC),
	}, {
		UserID:    "01020304-0506-0708-090a-0b0c0d0e0f10",
		Start:     time.Date(2020, 2, 1, 8, 0, 0, 0, time.UTC),
		End:       time.Date(2020, 2, 15, 8, 0, 0, 0, time.UTC),
		Truncated: true,
	}}
	generatedAt := time.Date(2020, 1, 1, 5, 0, 0, 0, time.UTC)
	r := renderData{
		ApplicationName: "GoAlert",
		ScheduleID:      uuid.MustParse("100f0e0d-0c0b-0a09-0807-060504030201"),
		ScheduleName:    "Sched",
		Shifts:          shifts,
		ReminderMinutes: []int{5, 10},
		Version:         "dev",
		GeneratedAt:     generatedAt,
	}
	iCal, err := r.renderICal()
	require.NoError(t, err)
	expected := strings.Join([]string{
		"BEGIN:VCALENDAR",
		"PRODID:-//GoAlert//dev//EN",
		"VERSION:2.0",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		"UID:4c7d37bf28d64eccc1e74a3889cfc97f6839a00fa781c91721058df915de27ce",
		"SUMMARY:On-Call (GoAlert: Sched)",
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
		"BEGIN:VEVENT",
		"UID:39514a8d78cada01fe7b66e1f9e511799cae36420a4b86a1d909c0bbbb99b747",
		"SUMMARY:On-Call (GoAlert: Sched) Begins*",
		"DESCRIPTION:The end time of this shift is unknown and will continue beyond what is displayed.",
		"DTSTAMP:20200101T050000Z",
		"DTSTART:20200201T080000Z",
		"DTEND:20200215T080000Z",
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
