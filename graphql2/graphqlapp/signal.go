package graphqlapp

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/search"
	"github.com/target/goalert/service"
	"github.com/target/goalert/service/rule"
	"github.com/target/goalert/signal"
	"github.com/target/goalert/validation/validate"
)

type Signal App

func (a *App) Signal() graphql2.SignalResolver { return (*Signal)(a) }

func (s *Signal) ID(ctx context.Context, raw *signal.Signal) (string, error) {
	return strconv.FormatInt(raw.ID, 10), nil
}

func (s *Signal) SignalID(ctx context.Context, raw *signal.Signal) (int, error) {
	return int(raw.ID), nil
}

func (s *Signal) OutgoingPayload(ctx context.Context, raw *signal.Signal) (string, error) {
	payloadBytes, err := json.Marshal(raw.OutgoingPayload)
	if err != nil {
		return "", err
	}
	return string(payloadBytes), nil
}

func (s *Signal) Service(ctx context.Context, raw *signal.Signal) (*service.Service, error) {
	return (*App)(s).FindOneService(ctx, raw.ServiceID)
}

func (s *Signal) ServiceRule(ctx context.Context, raw *signal.Signal) (*rule.Rule, error) {
	return s.ServiceRuleStore.FindOne(ctx, raw.ServiceRuleID)
}

func (q *Query) Signal(ctx context.Context, id int) (*signal.Signal, error) {
	return q.SignalStore.FindOne(ctx, id)
}

func (q *Query) Signals(ctx context.Context, opts *graphql2.SignalSearchOptions) (conn *graphql2.SignalConnection, err error) {
	if opts == nil {
		opts = &graphql2.SignalSearchOptions{}
	}

	var s signal.SearchOptions
	if opts.First != nil {
		s.Limit = *opts.First
	}

	if s.Limit == 0 {
		s.Limit = 15
	}
	s.Omit = opts.Omit

	err = validate.Many(
		validate.Range("ServiceIDs", len(opts.FilterByServiceID), 0, 50),
		validate.Range("ServiceRuleIDs", len(opts.FilterByServiceRuleID), 0, 50),
		validate.Range("First", s.Limit, 1, 1000),
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
		s.ServiceFilter.IDs = opts.FilterByServiceID
		if s.ServiceFilter.IDs != nil {
			s.ServiceFilter.Valid = true
		}

		s.ServiceRuleFilter.IDs = opts.FilterByServiceRuleID
		if s.ServiceRuleFilter.IDs != nil {
			s.ServiceRuleFilter.Valid = true
		}

		if opts.Sort != nil {
			switch *opts.Sort {
			case graphql2.SignalSearchSortDateID:
				s.Sort = signal.SortModeDateID
			case graphql2.SignalSearchSortDateIDReverse:
				s.Sort = signal.SortModeDateIDReverse
			}
		}
		if opts.CreatedBefore != nil {
			s.Before = *opts.CreatedBefore
		}
		if opts.NotCreatedBefore != nil {
			s.NotBefore = *opts.NotCreatedBefore
		}
	}

	s.Limit++

	signals, err := q.SignalStore.Search(ctx, &s)
	if err != nil {
		return conn, err
	}

	conn = new(graphql2.SignalConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(signals) == s.Limit {
		conn.PageInfo.HasNextPage = true
		signals = signals[:len(signals)-1]
	}
	conn.Nodes = signals
	if len(signals) > 0 {
		s.After.ID = int(conn.Nodes[len(conn.Nodes)-1].ID)
		s.After.Timestamp = conn.Nodes[len(conn.Nodes)-1].Timestamp
		cur, err := search.Cursor(s)
		if err != nil {
			return nil, errors.Wrap(err, "serialize cursor")
		}
		conn.PageInfo.EndCursor = &cur
	}

	return conn, nil
}
