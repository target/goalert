package calsub

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/version"
)

type iCalRenderData struct {
	ApplicationName string
	Shifts          []oncall.Shift
	ReminderMinutes []int
	Version         string
	GeneratedAt     time.Time
	EventUIDs       []string
}

// RFC can be found at https://tools.ietf.org/html/rfc5545
var iCalTemplate = template.Must(template.New("ical").Parse(strings.ReplaceAll(`BEGIN:VCALENDAR
PRODID:-//{{.ApplicationName}}//{{.Version}}//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
{{- $mins := .ReminderMinutes }}
{{- $genTime := .GeneratedAt }}
{{- $eventUIDs := .EventUIDs}}
{{- range $i, $s := .Shifts}}
BEGIN:VEVENT
UID:{{index $eventUIDs $i}}
SUMMARY:On-Call Shift{{if $s.Truncated}} Begins*
DESCRIPTION:The end time of this shift is unknown and will continue beyond what is displayed.
{{- end }}
DTSTAMP:{{$genTime.UTC.Format "20060102T150405Z"}}
DTSTART:{{.Start.UTC.Format "20060102T150405Z"}}
DTEND:{{.End.UTC.Format "20060102T150405Z"}}
{{- range $mins}}
BEGIN:VALARM
ACTION:DISPLAY
DESCRIPTION:REMINDER
TRIGGER:-PT{{.}}M
END:VALARM
{{- end}}
END:VEVENT
{{- end}}
END:VCALENDAR
`, "\n", "\r\n")))

func (cs Subscription) renderICalFromShifts(appName string, shifts []oncall.Shift, generatedAt time.Time) ([]byte, error) {
	var eventUIDs []string
	for _, s := range shifts {
		t := s.End
		if s.Truncated {
			t = s.Start
		}
		sum := sha256.Sum256([]byte(s.UserID + cs.ScheduleID + t.Format(time.RFC3339)))
		eventUIDs = append(eventUIDs, hex.EncodeToString(sum[:]))
	}
	data := iCalRenderData{
		ApplicationName: appName,
		Shifts:          shifts,
		ReminderMinutes: cs.Config.ReminderMinutes,
		Version:         version.GitVersion(),
		GeneratedAt:     generatedAt,
		EventUIDs:       eventUIDs,
	}
	buf := bytes.NewBuffer(nil)

	err := iCalTemplate.Execute(buf, data)
	if err != nil {
		return nil, errors.Wrap(err, "render ical template:")
	}

	return buf.Bytes(), nil
}
