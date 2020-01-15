package calendarsubscription

import (
	"net/http"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

// ServeICalData will return an iCal file for the subscription associated with the current request.
func (b *Store) ServeICalData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	src := permission.Source(ctx)
	if src.Type != permission.SourceTypeCalendarSubscription {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	cs, err := b.FindOne(ctx, src.ID)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	var n time.Time
	err = b.now.QueryRowContext(ctx).Scan(&n)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	shifts, err := b.oc.HistoryBySchedule(ctx, cs.ScheduleID, n, n.AddDate(0, 1, 0))
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	// filter out other users
	filtered := shifts[:0]
	for _, s := range shifts {
		if s.UserID != cs.UserID {
			continue
		}
		filtered = append(filtered, s)
	}

	calData := []byte("not implemented")
	// calData := renderICalFromShifts(shifts, cs.Config)

	w.Header().Set("Content-Type", "text/calendar")
	w.Write(calData)
}
