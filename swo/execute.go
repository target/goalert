package swo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func (m *Manager) SendProposal() (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *Manager) ProposalIsValid() (bool, error) {
	return false, nil
}

func (m *Manager) DoExecute(ctx context.Context) error {
	/*
		- initial sync
		- loop until few changes
		- send proposal
		- loop until proposal is valid
		- execute proposal

	*/

	return m.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
		m.Progressf(ctx, "scanning tables...")
		tables, err := ScanTables(ctx, oldConn)
		if err != nil {
			return fmt.Errorf("scan tables: %w", err)
		}

		m.Progressf(ctx, "enabling change log")
		err = EnableChangeLog(ctx, tables, oldConn)
		if err != nil {
			return fmt.Errorf("enable change log: %w", err)
		}

		m.Progressf(ctx, "disabling triggers")
		err = DisableTriggers(ctx, tables, newConn)
		if err != nil {
			return fmt.Errorf("disable triggers: %w", err)
		}

		m.Progressf(ctx, "performing initial sync")
		err = m.InitialSync(ctx, oldConn, newConn)
		if err != nil {
			return fmt.Errorf("initial sync: %w", err)
		}

		getTable := func(name string) *Table {
			for _, t := range tables {
				if t.Name == name {
					return &t
				}
			}
			return nil
		}

		for {
			// sync in a loop until DB is up-to-date
			n, err := LoopSync(ctx, getTable, oldConn, newConn)
			if err != nil {
				return fmt.Errorf("loop sync: %w", err)
			}
			m.Progressf(ctx, "sync: %d changes", n)
			if n == 0 {
				time.Sleep(time.Second)
			}
		}

		return errors.New("not implemented")
	})
}

// DisableTriggers will disable all triggers in the new DB.
func DisableTriggers(ctx context.Context, tables []Table, conn *pgx.Conn) error {
	for _, table := range tables {
		_, err := conn.Exec(ctx, fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER USER", table.QuotedName()))
		if err != nil {
			return fmt.Errorf("%s: %w", table.Name, err)
		}
	}

	return nil
}

func LoopSync(ctx context.Context, getTable func(string) *Table, oldConn, newConn *pgx.Conn) (int, error) {
	oldTx, newTx, err := syncTx(ctx, oldConn, newConn)
	if err != nil {
		return 0, fmt.Errorf("sync tx: %w", err)
	}
	defer oldTx.Rollback(ctx)
	defer newTx.Rollback(ctx)

	n, err := syncChangeLog(ctx, getTable, oldTx, newTx)
	if err != nil {
		return 0, fmt.Errorf("sync change log: %w", err)
	}

	err = newTx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit dst: %w", err)
	}

	err = oldTx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit src: %w", err)
	}

	return n, nil
}

func FinalSync(ctx context.Context, oldConn, newConn *pgx.Conn) error {
	return nil
}

func syncTx(ctx context.Context, oldConn, newConn *pgx.Conn) (old, new pgx.Tx, err error) {
	srcTx, err := oldConn.BeginTx(ctx, pgx.TxOptions{
		AccessMode:     pgx.ReadWrite,
		IsoLevel:       pgx.Serializable,
		DeferrableMode: pgx.Deferrable,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("begin src: %w", err)
	}

	dstTx, err := newConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		srcTx.Rollback(ctx)
		return nil, nil, fmt.Errorf("begin dst: %w", err)
	}

	return srcTx, dstTx, nil
}

const fetchQuery = `
with tx_max_id as (
	select max(id), tx_id
	from change_log
	group by tx_id
)
select id, op, table_name, row_id, row_data
from change_log c
join tx_max_id max_id on max_id.tx_id = c.tx_id
order by
	max_id.max,
	cmd_id::text::int
`

func syncChangeLog(ctx context.Context, getTable func(string) *Table, oldConn, newConn pgx.Tx) (int, error) {
	var b pgx.Batch
	var rowID, table, op string
	var data []byte
	var toDelete []int
	var id int
	_, err := oldConn.QueryFunc(ctx, fetchQuery, nil, []interface{}{&id, &op, &table, &rowID, &data}, func(pgx.QueryFuncRow) error {
		t := getTable(table)
		if t == nil {
			return fmt.Errorf("unknown table: %s", table)
		}

		switch op {
		case "INSERT":
			b.Queue(t.InsertOneRowQuery(), data)
		case "UPDATE":
			b.Queue(t.UpdateOneRowQuery(), rowID, data)
		case "DELETE":
			b.Queue(t.DeleteOneRowQuery(), rowID)
		default:
			return fmt.Errorf("unknown op: %s", op)
		}
		toDelete = append(toDelete, id)

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("query changes: %w", err)
	}

	err = newConn.SendBatch(ctx, &b).Close()
	if err != nil {
		return 0, fmt.Errorf("send batch: %w", err)
	}

	_, err = oldConn.Exec(ctx, "delete from change_log where id = any($1)", toDelete)

	return len(toDelete), err
}
