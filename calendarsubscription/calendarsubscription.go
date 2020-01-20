package calendarsubscription

import (
	"bytes"
	"html/template"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/validation/validate"
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

func (cs CalendarSubscription) renderICalFromShifts(shifts []oncall.Shift, reminderMinutes []int, scheduleName string) ([]byte, error) {
	type iCalOptions struct {
		Shifts          []oncall.Shift `json:"s,omitempty"`
		ReminderMinutes []int          `json:"r,omitempty"`
		ScheduleName    string         `json:"n,omitempty"`
	}

	var iCalTemplate = `
		BEGIN:VCALENDAR
		VERSION:2.0
		PRODID:-//ZContent.net//Zap Calendar 1.0//EN
		NAME:{{.ScheduleName}}
		CALSCALE:GREGORIAN
		METHOD:PUBLISH
		{{range .Shifts}}
		BEGIN:VEVENT
		SUMMARY:On-Call Shift
		DTSTART:{{.Start}}
		DTEND:{{.End}}
		END:VEVENT
		{{end}}

		{{range .ReminderMinutes}}
		BEGIN:VALARM
		ACTION:DISPLAY
		DESCRIPTION:REMINDER
		TRIGGER:-PT{{.}}M
		END:VALARM
		{{end}}

		END:VCALENDAR`

	iCal := template.Must(template.New("iCal").Parse(iCalTemplate))
	i := iCalOptions{shifts, reminderMinutes, scheduleName}

	buf := bytes.NewBuffer(nil)
	err := iCal.Execute(buf, i)
	if err != nil {
		return nil, errors.Wrap(err, "render ical template:")
	}

	return buf.Bytes(), nil
}
