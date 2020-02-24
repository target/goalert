package calendarsubscription

import (
	"bytes"
	"html/template"
	"strings"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/validation/validate"
	"github.com/target/goalert/version"
)

// CalendarSubscription stores the information from user subscriptions
type CalendarSubscription struct {
	ID         string
	Name       string
	UserID     string
	ScheduleID string
	LastAccess time.Time
	Disabled   bool

	// Config provides necessary parameters CalendarSubscription Config (i.e. ReminderMinutes)
	Config struct {
		ReminderMinutes []int
	}

	token string
}

type iCalRenderData struct {
	Shifts          []oncall.Shift
	ReminderMinutes []int
	Version         string
	GeneratedAt     time.Time
	EventIDs        []string
}

// RFC can be found at https://tools.ietf.org/html/rfc5545
var iCalTemplate = template.Must(template.New("ical").Parse(strings.ReplaceAll(`BEGIN:VCALENDAR
PRODID:-//GoAlert//{{.Version}}//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
{{- $mins := .ReminderMinutes }}
{{- $genTime := .GeneratedAt }}
{{- $eid := .EventIDs}}
{{- range $index, $element := .Shifts}}
BEGIN:VEVENT
UID:{{index $eid $index}}
SUMMARY:On-Call Shift
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

// Token returns the authorization token associated with this CalendarSubscription. It
// is only available when calling CreateTx.
func (cs CalendarSubscription) Token() string { return cs.token }

// Normalize will validate and produce a normalized CalendarSubscription struct.
func (cs CalendarSubscription) Normalize() (*CalendarSubscription, error) {
	if cs.ID == "" {
		cs.ID = uuid.NewV4().String()
	}

	err := validate.Many(
		validate.Range("ReminderMinutes", len(cs.Config.ReminderMinutes), 0, 15),
		validate.IDName("Name", cs.Name),
		validate.UUID("ID", cs.ID),
		validate.UUID("UserID", cs.UserID),
	)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

func (cs CalendarSubscription) renderICalFromShifts(shifts []oncall.Shift, generatedAt time.Time) ([]byte, error) {
	var ids []string
	for i := 0; i < len(shifts); i++ {
		ids = append(ids, uuid.NewV4().String())
	}
	data := iCalRenderData{Shifts: shifts, ReminderMinutes: cs.Config.ReminderMinutes, Version: version.GitVersion(), GeneratedAt: generatedAt, EventIDs: ids}
	buf := bytes.NewBuffer(nil)

	err := iCalTemplate.Execute(buf, data)
	if err != nil {
		return nil, errors.Wrap(err, "render ical template:")
	}

	return buf.Bytes(), nil
}
