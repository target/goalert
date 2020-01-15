package calendarsubscription

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
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
}

type post struct {
	SubscriptionID string `json:"s"`
}

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

func clientError(w http.ResponseWriter, code int, err error) bool {
	if err == nil {
		return false
	}

	http.Error(w, http.StatusText(code), code)
	return true
}

// CalendarSubscription
func CSToEventsAPI(oDB oncall.Store, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var g post
		err := json.NewDecoder(r.Body).Decode(&g)
		if clientError(w, http.StatusBadRequest, err) {
			log.Logf(ctx, "bad request from calendarsubscription: %v", err)
			return
		}

		ctx = log.WithFields(ctx, log.Fields{
			"SubscriptionID": g.SubscriptionID,
		})

		err = validate.UUID("SubscriptionID", g.SubscriptionID)
		if err != nil {
			// todo
		}

		s, err := NewStore(ctx, db)
		// get calendar subscription details
		_, err = s.FindOne(ctx, g.SubscriptionID)
		if err != nil {
			/// todo
		}

		// get shifts
		t1, _ := time.Parse(time.RFC3339, "2020-01-01T22:08:41+00:00")
		t2, _ := time.Parse(time.RFC3339, "2020-01-07T22:08:41+00:00")

		shifts, err := oDB.HistoryBySchedule(ctx, "59aea4b0-75f0-4af3-9824-644abf8dd29a", t1, t2)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var uShifts []oncall.Shift
		for _, s := range shifts {
			if s.UserID == "cb75f78a-0f7c-42fa-99f8-6b30e92a9518" {
				uShifts = append(uShifts, s)
			}
		}

		// get iCal data
		_, err = ICal(uShifts, t1, t2, false)

		if errutil.HTTPError(ctx, w, errors.Wrap(err, "serve iCalendar")) {
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}
