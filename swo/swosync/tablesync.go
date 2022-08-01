package swosync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swoinfo"
	"github.com/target/goalert/util/sqlutil"
)

// TableSync is a helper for syncing tables from the source database to the target database.
type TableSync struct {
	tables []swoinfo.Table

	changes []changeEntry

	changedTables []string
	changedData   map[changeID]json.RawMessage
}

type changeEntry struct {
	id int64
	changeID
}

// NewTableSync creates a new TableSync for the given tables.
func NewTableSync(tables []swoinfo.Table) *TableSync {
	return &TableSync{
		tables:      tables,
		changedData: make(map[changeID]json.RawMessage),
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

		c.changes = append(c.changes, changeEntry{id: id, changeID: changeID{table, rowID}})
	}

	return rows.Err()
}

// HasChanges returns true after ScanBatchChangeRead has been called, if there are changes.
func (c *TableSync) HasChanges() bool { return len(c.changes) > 0 }

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
	rowIDsByTable := make(map[string][]string)
	for _, chg := range c.changes {
		rowIDsByTable[chg.Table] = append(rowIDsByTable[chg.Table], chg.Row)
	}

	for _, table := range c.tables {
		rowIDs := rowIDsByTable[table.Name()]
		if len(rowIDs) == 0 {
			continue
		}

		c.changedTables = append(c.changedTables, table.Name())
		arg, cast := castIDs(table, rowIDs)
		b.Queue(fmt.Sprintf(`select id::text, to_jsonb(row) from %s row where id%s = any($1)`, sqlutil.QuoteID(table.Name()), cast), arg)
	}
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
	if len(c.changes) == 0 {
		return 0, nil
	}

	var ids []int
	for _, chg := range c.changes {
		ids = append(ids, int(chg.id))
	}
	_, err := srcConn.Exec(ctx, `delete from change_log where id = any($1)`, sqlutil.IntArray(ids))
	if err != nil {
		return 0, fmt.Errorf("delete %d change log rows: %w", len(ids), err)
	}

	return int64(len(ids)), nil
}

func (c *TableSync) AddBatchWrites(b *pgx.Batch, dstRows rowSet) {
	type pending struct {
		inserts []json.RawMessage
		updates []json.RawMessage
		deletes []string
	}
	pendingByTable := make(map[string]*pending)
	for _, chg := range c.changes {
		p := pendingByTable[chg.Table]
		if p == nil {
			p = &pending{}
			pendingByTable[chg.Table] = p
		}
		newRowData := c.changedData[chg.changeID]
		if newRowData == nil {
			// row was deleted
			dstRows.Delete(chg.changeID)
			p.deletes = append(p.deletes, chg.Row)
			continue
		}

		if dstRows.Has(chg.changeID) {
			// row was updated
			p.updates = append(p.updates, newRowData)
		} else {
			// row was inserted
			dstRows.Set(chg.changeID)
			p.inserts = append(p.inserts, newRowData)
		}
	}

	// insert, then update, then reverse delete
	for _, t := range c.tables {
		p := pendingByTable[t.Name()]
		if p == nil || len(p.inserts) == 0 {
			continue
		}
		b.Queue(insertRowsQuery(t), p.inserts)
	}

	for _, t := range c.tables {
		p := pendingByTable[t.Name()]
		if p == nil || len(p.updates) == 0 {
			continue
		}
		b.Queue(updateRowsQuery(t), p.updates)
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

func updateRowsQuery(t swoinfo.Table) string {
	var s strings.Builder
	fmt.Fprintf(&s, "update %s dst\n", sqlutil.QuoteID(t.Name()))
	fmt.Fprintf(&s, "set ")
	for i, col := range t.Columns() {
		if i > 0 {
			fmt.Fprintf(&s, ", ")
		}
		fmt.Fprintf(&s, "%s = data.%s", sqlutil.QuoteID(col), sqlutil.QuoteID(col))
	}
	fmt.Fprintf(&s, "\nfrom json_populate_recordset(null::%s, $1) as data\n", sqlutil.QuoteID(t.Name()))
	fmt.Fprintf(&s, "where dst.id = data.id")

	return s.String()
}
