package swosync

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/util/sqlutil"
)

// SequenceSync is a helper for synchronizing sequences.
type SequenceSync struct {
	names     []string
	lastValue []int64
	isCalled  []bool
}

// NewSequenceSync creates a new SequenceSync for the given sequence names.
func NewSequenceSync(names []string) *SequenceSync {
	return &SequenceSync{names: names}
}

// AddBatchReads queues up select statements to retrieve the current values of the sequences.
func (s *SequenceSync) AddBatchReads(b *pgx.Batch) {
	for _, seqName := range s.names {
		b.Queue("select last_value, is_called from " + sqlutil.QuoteID(seqName))
	}
}

// ScanReads scans the results of the added batch reads.
func (s *SequenceSync) ScanBatchReads(res pgx.BatchResults) error {
	for _, seqName := range s.names {
		var last int64
		var isCalled bool
		err := res.QueryRow().Scan(&last, &isCalled)
		if err != nil {
			return fmt.Errorf("read changes: scan seq %s: %w", seqName, err)
		}
		s.lastValue = append(s.lastValue, last)
		s.isCalled = append(s.isCalled, isCalled)
	}

	return nil
}

// AddBatchWrites queues up update statements to set the current values of the sequences.
func (s *SequenceSync) AddBatchWrites(b *pgx.Batch) {
	for i, seqName := range s.names {
		b.Queue("select pg_catalog.setval($1, $2, $3)", seqName, s.lastValue[i], s.isCalled[i])
	}
}
