package signal

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SortMode indicates the mode of sorting for signals.
type SortMode int

const (
	// SortModeDateID will sort signals by date newest first, falling back to ID (newest/highest first)
	SortModeDateID SortMode = iota

	// SortModeDateIDReverse will sort signals by date oldest first, falling back to ID (oldest/lowest first)
	SortModeDateIDReverse
)

// SearchOptions contains criteria for filtering and sorting signals.
type SearchOptions struct {

	// ServiceFilter, if specified, will restrict signals to those with a matching ServiceID on IDs, if valid.
	ServiceFilter IDFilter `json:"v,omitempty"`

	// ServiceRuleFilter, if specified, will restrict signals to those with a matching ServiceRuleID on IDs, if valid.
	ServiceRuleFilter IDFilter `json:"r,omitempty"`

	After SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of signal IDs to exclude from the results.
	Omit []int `json:"o,omitempty"`

	// Limit restricts the maximum number of rows returned. Default is 50.
	// Note: Limit is applied AFTER AfterID is taken into account.
	Limit int `json:"-"`

	// Sort allows customizing the sort method.
	Sort SortMode `json:"z,omitempty"`

	// NotBefore will omit any signals with a timestamp before the provided time.
	NotBefore time.Time `json:"n,omitempty"`

	// Before will only include signals with a timestamp before the provided time.
	Before time.Time `json:"b,omitempty"`
}

type IDFilter struct {
	Valid bool     `json:"v,omitempty"`
	IDs   []string `json:"i,omitempty"`
}

type SearchCursor struct {
	ID        int       `json:"i,omitempty"`
	Timestamp time.Time `json:"c,omitempty"`
}

func (opts SearchOptions) Normalize() (*SearchOptions, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.Range("Limit", opts.Limit, 0, 1001),
		validate.Range("Omit", len(opts.Omit), 0, 50),
		validate.OneOf("Sort", opts.Sort, SortModeDateID, SortModeDateIDReverse),
	)
	if err != nil {
		return nil, err
	}

	return &opts, err
}

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]Signal, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = new(SearchOptions)
	}

	opts, err = opts.Normalize()
	if err != nil {
		return nil, err
	}

	params := gadb.SignalSearchParams{SortMode: int32(opts.Sort), Limit: int32(opts.Limit)}
	params.Omit = make([]int64, len(opts.Omit))
	for i, id := range opts.Omit {
		params.Omit[i] = int64(id)
	}
	if opts.ServiceFilter.Valid {
		params.AnyServiceID, err = validate.ParseManyUUID("ServiceIDs", opts.ServiceFilter.IDs, 50)
		if err != nil {
			return nil, err
		}
	}
	if opts.ServiceRuleFilter.Valid {
		params.AnyServiceRuleID, err = validate.ParseManyUUID("ServiceRuleIDs", opts.ServiceRuleFilter.IDs, 50)
		if err != nil {
			return nil, err
		}
	}
	if !opts.Before.IsZero() {
		params.BeforeTime = sql.NullTime{Valid: true, Time: opts.Before}
	}
	if !opts.NotBefore.IsZero() {
		params.NotBeforeTime = sql.NullTime{Valid: true, Time: opts.NotBefore}
	}
	if opts.After.ID != 0 {
		params.AfterID = sql.NullInt64{Valid: true, Int64: int64(opts.After.ID)}
		params.AfterTimestamp = opts.After.Timestamp
	}

	rows, err := gadb.New(s.db).SignalSearch(ctx, params)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}

	signals := make([]Signal, 0, opts.Limit)
	for _, row := range rows {
		payload := make(map[string]interface{})
		err := json.Unmarshal(row.OutgoingPayload, &payload)
		if err != nil {
			log.Log(log.WithField(ctx, "SignalID", row.ID), errors.Wrap(err, "unmarshal signal payload"))
		}
		signals = append(signals, Signal{
			ID:              row.ID,
			ServiceID:       row.ServiceID.String(),
			ServiceRuleID:   row.ServiceRuleID.String(),
			OutgoingPayload: payload,
			Scheduled:       row.Scheduled,
			Timestamp:       row.Timestamp,
		})
	}

	return signals, nil
}
