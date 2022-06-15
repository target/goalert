package alertmetrics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
)

type Store struct {
	db *sql.DB

	findMetrics *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,

		findMetrics: p.P(`select alert_id, coalesce(time_to_ack, time_to_close), time_to_close, escalated, closed_at from alert_metrics where alert_id = any($1)`),
	}, p.Err
}

type Metric struct {
	ID          int
	TimeToAck   time.Duration
	TimeToClose time.Duration
	ClosedAt    time.Time
	Escalated   bool
}

func (s *Store) FindMetrics(ctx context.Context, alertIDs []int) ([]Metric, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	rows, err := s.findMetrics.QueryContext(ctx, sqlutil.IntArray(alertIDs))
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var m Metric
		var ack, cls pgtype.Interval
		if err := rows.Scan(&m.ID, &ack, &cls, &m.Escalated, &m.ClosedAt); err != nil {
			return nil, fmt.Errorf("scanning metric: %w", err)
		}
		err = ack.AssignTo(&m.TimeToAck)
		if err != nil {
			return nil, fmt.Errorf("assigning metric: %w", err)
		}
		err = cls.AssignTo(&m.TimeToClose)
		if err != nil {
			return nil, fmt.Errorf("assigning metric: %w", err)
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}
