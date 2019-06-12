package graphqlapp

import (
	"context"
	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/log"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"
)

type AlertLog App

func (a *App) AlertLog() graphql2.AlertLogResolver { return (*AlertLog)(a) }

func (l *AlertLog) Event(ctx context.Context, log *alert.Log) (graphql2.Event, error) {
	return graphql2.Event(log.Event), nil
}

func (l *AlertLog) AlertID(ctx context.Context, log *alert.Log) (int, error) {
	var s alertlog.SearchOptions
	s.Start = log.Timestamp
	s.Event = alertlog.Type(log.Event)
	s.Limit = 1

	entries, _, err := l.LogStore.Search(ctx, &s)
	if err != nil {
		return -1, err
	}
	return entries[0].AlertID(), nil
}

func (q *Query) Alertlogs(ctx context.Context, opts *graphql2.AlertLogSearchOptions) (conn *graphql2.AlertLogConnection, err error) {
	if opts == nil {
		opts = &graphql2.AlertLogSearchOptions{}
	}

	var s alertlog.SearchOptions

	if opts.After != nil && *opts.After != "" {
		err = search.ParseCursor(*opts.After, &s)
		if err != nil {
			return nil, errors.Wrap(err, "parse cursor")
		}
	}

	if opts.FilterByAlertID != nil {
		s.AlertID = *opts.FilterByAlertID
	} else {
		s.AlertID = 0
	}

	if opts.FilterByServiceID != nil {
		s.ServiceID = *opts.FilterByServiceID
	}

	s.SortDesc = *opts.SortDesc

	if opts.Event != nil {
		switch *opts.Event {
		case graphql2.EventCreated:
			s.Event = "created"
		case graphql2.EventClosed:
			s.Event = "closed"
		case graphql2.EventNotificationSent:
			s.Event = "notification_sent"
		case graphql2.EventEscalated:
			s.Event = "escalated"
		case graphql2.EventAcknowledged:
			s.Event = "acknowledged"
		case graphql2.EventPolicyUpdated:
			s.Event = "policy_updated"
		case graphql2.EventDuplicateSuppressed:
			s.Event = "duplicate_suppressed"
		case graphql2.EventEscalationRequest:
			s.Event = "escalation_request"
		}
	}

	if opts.First != nil {
		s.Limit = *opts.First
	}
	if s.Limit == 0 {
		s.Limit = 15
	}

	err = validate.Many(
		validate.Range("First", s.Limit, 1, 100),
	)
	if err != nil {
		return nil, err
	}

	s.Limit++

	entries, _, err := q.LogStore.Search(ctx, &s)
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.AlertLogConnection)
	if len(entries) == s.Limit {
		entries = entries[:len(entries)-1]
		conn.PageInfo.HasNextPage = true
	}

	if len(entries) > 0 {
		last := entries[len(entries)-1]
		s.After.ID = last.ID()
		cur, err := search.Cursor(s)
		if err != nil {
			return nil, errors.Wrap(err, "serialize cursor")
		}
		conn.PageInfo.EndCursor = &cur
	}

	logs := make([]alert.Log, len(entries))
	for i, e := range entries {
		var l alert.Log
		l.Timestamp = e.Timestamp()
		l.Message = e.String()
		l.Event = alert.LogEvent(e.Type())
		logs[i] = l
	}

	conn.Nodes = logs
	return conn, err
}
