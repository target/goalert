package calsub

import (
	"net/http"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

// ServeICalData will return an iCal file for the subscription associated with the current request.
func (s *Store) ServeICalData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	src := permission.Source(ctx)
	cfg := config.FromContext(ctx)
	if src.Type != permission.SourceTypeCalendarSubscription || cfg.General.DisableCalendarSubscriptions {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	cs, err := s.FindOne(ctx, src.ID)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	n, err := gadb.New(s.db).Now(ctx)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	shifts, err := s.oc.HistoryBySchedule(ctx, cs.ScheduleID, n, n.AddDate(1, 0, 0))
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

	calData, err := cs.renderICalFromShifts(cfg.ApplicationName(), filtered, n)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	w.Header().Set("Content-Type", "text/calendar")
	_, _ = w.Write(calData)
}
