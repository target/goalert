package swosync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/swo/swoinfo"
	"github.com/target/goalert/util/sqlutil"
)

// TableSync is a helper for syncing tables from the source database to the target database.
type TableSync struct {
	tables []swoinfo.Table

	changedTables []string
	changedRowIDs map[string][]string
	changedData   map[changeID]json.RawMessage
	changeLogIDs  []int
}

type changeID struct{ Table, Row string }

// NewTableSync creates a new TableSync for the given tables.
func NewTableSync(tables []swoinfo.Table) *TableSync {
	return &TableSync{
		tables:        tables,
		changedData:   make(map[changeID]json.RawMessage),
		changedRowIDs: make(map[string][]string),
	}
}

// AddBatchChangeRead adds a query to the batch to read the changes from the source database.
func (c *TableSync) AddBatchChangeRead(b *pgx.Batch) {
	b.Queue(`select id, table_name, row_id from change_log`)
}

// ScanBatchChangeRead scans the results of the change read query.
func (c *TableSync) ScanBatchChangeRead(res pgx.BatchResults) error {
	rows, err := res.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var table string
		var rowID string
		if err := rows.Scan(&id, &table, &rowID); err != nil {
			return err
		}
		c.changeLogIDs = append(c.changeLogIDs, int(id))
		c.changedData[changeID{table, rowID}] = nil // mark as changed
		c.changedRowIDs[table] = append(c.changedRowIDs[table], rowID)
	}

	return rows.Err()
}

// HasChanges returns true after ScanBatchChangeRead has been called, if there are changes.
func (c *TableSync) HasChanges() bool { return len(c.changeLogIDs) > 0 }

func intIDs(ids []string) []int {
	var ints []int
	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			panic(err)
		}
		ints = append(ints, i)
	}
	return ints
}

// AddBatchRowReads adds a query to the batch to read all changed rows from the source database.
func (c *TableSync) AddBatchRowReads(b *pgx.Batch) {
	for _, table := range c.tables {
		rowIDs := unique(c.changedRowIDs[table.Name()])
		if len(rowIDs) == 0 {
			continue
		}

		c.changedTables = append(c.changedTables, table.Name())
		arg, cast := castIDs(table, rowIDs)
		b.Queue(fmt.Sprintf(`select id::text, to_jsonb(row) from %s row where id%s = any($1)`, sqlutil.QuoteID(table.Name()), cast), arg)
	}
}

func unique(ids []string) []string {
	sort.Strings(ids)

	uniq := ids[:0]
	var last string
	for _, id := range ids {
		if id == last {
			continue
		}
		uniq = append(uniq, id)
		last = id
	}
	return uniq
}

func castIDs(t swoinfo.Table, rowIDs []string) (interface{}, string) {
	var cast string
	switch t.IDType() {
	case "integer", "bigint":
		return sqlutil.IntArray(intIDs(rowIDs)), ""
	case "uuid":
		return sqlutil.UUIDArray(rowIDs), ""
	default:
		// anything else/unknown should be cast to text and compared to the string version
		// this is slower, but should only happen for small tables where the id column is an enum
		cast = "::text"
		fallthrough
	case "text":
		return sqlutil.StringArray(rowIDs), cast
	}
}

// ScanBatchRowReads scans the results of the row read queries.
func (c *TableSync) ScanBatchRowReads(res pgx.BatchResults) error {
	if len(c.changedTables) == 0 {
		return nil
	}

	for _, tableName := range c.changedTables {
		rows, err := res.Query()
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("query changed rows from %s: %w", tableName, err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			var row json.RawMessage
			if err := rows.Scan(&id, &row); err != nil {
				return fmt.Errorf("scan changed rows from %s: %w", tableName, err)
			}

			c.changedData[changeID{tableName, id}] = row
		}
	}

	return nil
}

// ExecDeleteChanges executes a query to deleted the change_log entries from the source database.
func (c *TableSync) ExecDeleteChanges(ctx context.Context, srcConn *pgx.Conn) (int64, error) {
	if len(c.changeLogIDs) == 0 {
		return 0, nil
	}

	_, err := srcConn.Exec(ctx, `delete from change_log where id = any($1)`, sqlutil.IntArray(c.changeLogIDs))
	if err != nil {
		return 0, fmt.Errorf("delete %d change log rows: %w", len(c.changeLogIDs), err)
	}

	return int64(len(c.changeLogIDs)), nil
}

func (c *TableSync) AddBatchWrites(b *pgx.Batch) {
	type pending struct {
		upserts []json.RawMessage
		deletes []string
	}
	pendingByTable := make(map[string]*pending)
	for id, data := range c.changedData {
		p := pendingByTable[id.Table]
		if p == nil {
			p = &pending{}
			pendingByTable[id.Table] = p
		}

		if data == nil {
			// row was deleted
			p.deletes = append(p.deletes, id.Row)
			continue
		}

		p.upserts = append(p.upserts, data)
	}

	// insert, then update, then reverse delete
	for _, t := range c.tables {
		p := pendingByTable[t.Name()]
		if p == nil || len(p.upserts) == 0 {
			continue
		}
		switch t.Name() {
		case "ep_step_on_call_users", "schedule_on_call_users":
			// due to unique constraint on shifts, we need to sort shift ends before new shifts
			sortOnCallData(p.upserts)
		}
		b.Queue(t.InsertJSONRowsQuery(true), p.upserts)
	}

	for i := range c.tables {
		// reverse-order tables
		t := c.tables[len(c.tables)-i-1]
		p := pendingByTable[t.Name()]
		if p == nil || len(p.deletes) == 0 {
			continue
		}
		arg, cast := castIDs(t, p.deletes)
		b.Queue(fmt.Sprintf(`delete from %s where id%s = any($1)`, sqlutil.QuoteID(t.Name()), cast), arg)
	}
}

// sortOnCallData sorts entries with a non-nil end time before entries with a nil end time
func sortOnCallData(data []json.RawMessage) {
	type onCallData struct {
		End *time.Time `json:"end_time"`
	}
	sort.Slice(data, func(i, j int) bool {
		var a, b onCallData
		if err := json.Unmarshal(data[i], &a); err != nil {
			panic(err)
		}
		if err := json.Unmarshal(data[j], &b); err != nil {
			panic(err)
		}
		return a.End != nil && b.End == nil
	})
}
