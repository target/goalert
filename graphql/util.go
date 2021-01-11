package graphql

import (
	"context"
	"database/sql"
	"time"

	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"

	g "github.com/graphql-go/graphql"
	"github.com/pkg/errors"
)

// ISOTimestamp is a timestamp formatted as a string in the ISO format
var ISOTimestamp = g.NewScalar(g.ScalarConfig{
	Name:        "ISOTimestamp",
	Description: "ISOTimestamp is a timestamp formatted as a string in the ISO format (RFC3339).",
	Serialize: func(val interface{}) interface{} {
		return val.(time.Time).Format(time.RFC3339Nano)
	},
})

// HourTime is a timestamp containing only the hour and minute
var HourTime = g.NewScalar(g.ScalarConfig{
	Name:        "HourTime",
	Description: "HourTime is a timestamp containing only the hour and minute.",
	Serialize: func(val interface{}) interface{} {
		return val.(rule.Clock).String()
	},
})

type scrubber struct{ ctx context.Context }

func isCtxCause(err error) bool {
	if errors.Is(err, context.Canceled) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, sql.ErrTxDone) {
		return true
	}

	// 57014 = query_canceled
	// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
	if e := sqlutil.MapError(err); e != nil && e.Code == "57014" {
		return true
	}

	return false
}

func newScrubber(ctx context.Context) *scrubber { return &scrubber{ctx: ctx} }
func (s *scrubber) scrub(val interface{}, err error) (interface{}, error) {
	if err == nil {
		return val, nil
	}
	if errors.Is(err, sql.ErrNoRows) || (s.ctx.Err() != nil && isCtxCause(err)) {
		log.Debug(s.ctx, errors.Wrap(err, "graphql"))
		return nil, nil
	}
	err = errutil.MapDBError(err)
	orig := err
	scrubbed, err := errutil.ScrubError(err)
	if scrubbed {
		log.Log(s.ctx, errors.Wrap(orig, "graphql"))
	} else {
		log.Debug(s.ctx, errors.Wrap(err, "graphql"))
	}
	return nil, err
}
