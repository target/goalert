package graphqlapp

import (
	context "context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/service"
	"github.com/target/goalert/validation/validate"
)

type Alert App
type AlertLogEntry App

func (a *App) Alert() graphql2.AlertResolver { return (*Alert)(a) }

func (a *App) AlertLogEntry() graphql2.AlertLogEntryResolver { return (*AlertLogEntry)(a) }

func (a *AlertLogEntry) ID(ctx context.Context, obj *alertlog.Entry) (int, error) {
	e := *obj
	return e.ID(), nil
}

func (a *AlertLogEntry) Timestamp(ctx context.Context, obj *alertlog.Entry) (*time.Time, error) {
	e := *obj
	t := e.Timestamp()
	return &t, nil
}

func (a *AlertLogEntry) Message(ctx context.Context, obj *alertlog.Entry) (string, error) {
	e := *obj
	return e.String(), nil
}

func (q *Query) Alert(ctx context.Context, alertID int) (*alert.Alert, error) {
	return (*App)(q).FindOneAlert(ctx, alertID)
}

func (q *Query) mergeFavorites(ctx context.Context, svcs []string) ([]string, error) {
	targets, err := q.FavoriteStore.FindAll(ctx, permission.UserID(ctx), []assignment.TargetType{assignment.TargetTypeService})
	if err != nil {
		return nil, err
	}
	if len(svcs) == 0 {
		for _, t := range targets {
			svcs = append(svcs, t.TargetID())
		}
	} else {
		// favorites AND serviceIDs
		m := make(map[string]bool, len(svcs))
		for _, o := range svcs {
			m[o] = true
		}
		// empty slice
		svcs = svcs[:0]

		for _, t := range targets {
			if !m[t.TargetID()] {
				continue
			}
			svcs = append(svcs, t.TargetID())
		}
		// Here we have the intersection of favorites and user-specified serviceIDs in opts.FilterByServiceID
	}
	return svcs, nil
}

func (q *Query) Alerts(ctx context.Context, opts *graphql2.AlertSearchOptions) (conn *graphql2.AlertConnection, err error) {
	if opts == nil {
		opts = new(graphql2.AlertSearchOptions)
	}

	var s alert.SearchOptions
	if opts.First != nil {
		s.Limit = *opts.First
	}

	if s.Limit == 0 {
		s.Limit = 15
	}
	if opts.Search != nil {
		s.Search = *opts.Search
	}
	s.Omit = opts.Omit

	err = validate.Many(
		validate.Range("ServiceIDs", len(opts.FilterByServiceID), 0, 50),
		validate.Range("First", s.Limit, 1, 100),
	)
	if err != nil {
		return nil, err
	}

	hasCursor := opts.After != nil && *opts.After != ""

	if hasCursor {
		err = search.ParseCursor(*opts.After, &s)
		if err != nil {
			return nil, errors.Wrap(err, "parse cursor")
		}
	} else {
		if opts.FavoritesOnly != nil && *opts.FavoritesOnly {
			s.Services, err = q.mergeFavorites(ctx, opts.FilterByServiceID)
			if err != nil {
				return nil, err
			}
		} else {
			s.Services = opts.FilterByServiceID
		}
		for _, f := range opts.FilterByStatus {
			switch f {
			case graphql2.AlertStatusStatusAcknowledged:
				s.Status = append(s.Status, alert.StatusActive)
			case graphql2.AlertStatusStatusUnacknowledged:
				s.Status = append(s.Status, alert.StatusTriggered)
			case graphql2.AlertStatusStatusClosed:
				s.Status = append(s.Status, alert.StatusClosed)
			}
		}
	}

	s.Limit++

	alerts, err := q.AlertStore.Search(ctx, &s)
	if err != nil {
		return conn, err
	}

	conn = new(graphql2.AlertConnection)
	if len(alerts) == s.Limit {
		conn.PageInfo.HasNextPage = true
		alerts = alerts[:len(alerts)-1]
	}
	conn.Nodes = alerts
	if len(alerts) > 0 {
		s.After.ID = conn.Nodes[len(conn.Nodes)-1].ID
		s.After.Status = conn.Nodes[len(conn.Nodes)-1].Status
		cur, err := search.Cursor(s)
		if err != nil {
			return nil, errors.Wrap(err, "serialize cursor")
		}
		conn.PageInfo.EndCursor = &cur
	}

	return conn, nil
}

func (a *Alert) ID(ctx context.Context, raw *alert.Alert) (string, error) {
	return fmt.Sprintf("Alert(%d)", raw.ID), nil
}
func (a *Alert) Status(ctx context.Context, raw *alert.Alert) (graphql2.AlertStatus, error) {
	switch raw.Status {
	case alert.StatusTriggered:
		return graphql2.AlertStatusStatusUnacknowledged, nil
	case alert.StatusClosed:
		return graphql2.AlertStatusStatusClosed, nil
	case alert.StatusActive:
		return graphql2.AlertStatusStatusAcknowledged, nil
	}
	return "", errors.New("unknown alert status " + string(raw.Status))
}
func (a *Alert) AlertID(ctx context.Context, raw *alert.Alert) (int, error) {
	return raw.ID, nil
}

func (a *Alert) State(ctx context.Context, raw *alert.Alert) (*alert.State, error) {
	return (*App)(a).FindOneAlertState(ctx, raw.ID)
}

func (a *Alert) Service(ctx context.Context, raw *alert.Alert) (*service.Service, error) {
	return (*App)(a).FindOneService(ctx, raw.ServiceID)
}

func (m *Mutation) CreateAlert(ctx context.Context, input graphql2.CreateAlertInput) (*alert.Alert, error) {
	// An alert when created will always have triggered status
	a := &alert.Alert{
		ServiceID: input.ServiceID,
		Summary:   input.Summary,
		Status:    alert.StatusTriggered,
	}

	if input.Details != nil {
		a.Details = *input.Details
	}

	return m.AlertStore.Create(ctx, a)
}

func (a *Alert) RecentEvents(ctx context.Context, obj *alert.Alert, opts *graphql2.AlertRecentEventsOptions) (*graphql2.AlertLogEntryConnection, error) {
	if opts == nil {
		opts = new(graphql2.AlertRecentEventsOptions)
	}

	var s alertlog.SearchOptions
	s.FilterAlertIDs = append(s.FilterAlertIDs, obj.ID)

	if opts.After != nil && *opts.After != "" {
		err := search.ParseCursor(*opts.After, &s)
		if err != nil {
			return nil, err
		}
	}

	if opts.Limit != nil {
		s.Limit = *opts.Limit
	}
	if s.Limit == 0 {
		s.Limit = search.DefaultMaxResults
	}

	s.Limit++

	logs, err := a.AlertLogStore.Search(ctx, &s)
	if err != nil {
		return nil, err
	}
	conn := new(graphql2.AlertLogEntryConnection)
	if len(logs) == s.Limit {
		logs = logs[:len(logs)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(logs) > 0 {
		last := logs[len(logs)-1]
		s.After.ID = last.ID()
		cur, err := search.Cursor(s)
		if err != nil {
			return nil, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = logs
	return conn, err
}

func (m *Mutation) EscalateAlerts(ctx context.Context, ids []int) ([]alert.Alert, error) {
	ids, err := m.AlertStore.EscalateMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	return m.AlertStore.FindMany(ctx, ids)
}

func (m *Mutation) UpdateAlerts(ctx context.Context, args graphql2.UpdateAlertsInput) ([]alert.Alert, error) {
	var status alert.Status

	err := validate.OneOf("Status", args.NewStatus, graphql2.AlertStatusStatusAcknowledged, graphql2.AlertStatusStatusClosed)
	if err != nil {
		return nil, err
	}

	switch args.NewStatus {
	case graphql2.AlertStatusStatusAcknowledged:
		status = alert.StatusActive
	case graphql2.AlertStatusStatusClosed:
		status = alert.StatusClosed
	}

	var updatedIDs []int
	updatedIDs, err = m.AlertStore.UpdateManyAlertStatus(ctx, status, args.AlertIDs)
	if err != nil {
		return nil, err
	}

	return m.AlertStore.FindMany(ctx, updatedIDs)
}
