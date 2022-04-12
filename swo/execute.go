package swo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/swo/swogrp"
	"github.com/target/goalert/util/sqlutil"
)

type Execute struct {
	err    error
	tables []Table

	seqNames []string

	mainDBConn, nextDBConn *pgx.Conn

	grp *swogrp.Group

	rowIDs map[string]map[string]struct{}

	stagedInserts []stagedID
	stagedDeletes []stagedID
}

func NewExecute(ctx context.Context, mainDBConn, nextDBConn *pgx.Conn, grp *swogrp.Group) (*Execute, error) {
	tables, err := ScanTables(ctx, mainDBConn)
	if err != nil {
		return nil, fmt.Errorf("scan tables: %w", err)
	}

	var seqNames []string
	var name string
	_, err = mainDBConn.QueryFunc(ctx, `
		select sequence_name
		from information_schema.sequences
		where
			sequence_catalog = current_database() and
			sequence_schema = 'public'
	`, nil, []interface{}{&name}, func(r pgx.QueryFuncRow) error {
		if name == "change_log_id_seq" {
			// skip, as it does not exist in next db
			return nil
		}
		seqNames = append(seqNames, name)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan sequences: %w", err)
	}

	return &Execute{
		tables:     tables,
		seqNames:   seqNames,
		mainDBConn: mainDBConn,
		nextDBConn: nextDBConn,
		grp:        grp,
	}, nil
}

func (m *Manager) DoExecute(ctx context.Context) error {
	return m.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
		exec, err := NewExecute(ctx, oldConn, newConn, m.grp)
		if err != nil {
			return err
		}

		exec.EnableChangeLog(ctx)
		exec.DisableNextDBTriggers(ctx)
		exec.WaitForActiveTx(ctx)
		exec.SyncFull(ctx)
		exec.ReadRowIDs(ctx)
		exec.SyncLoop(ctx)
		exec.PauseApps(ctx)
		exec.FinalSync(ctx)

		return exec.readErr()
	})
}

// PauseApps puts all nodes into a "paused" state:
// - Engine no longer cycles
// - Idle DB connections are disabled
// - Event listeners (postgres pub/sub) are disabled
func (e *Execute) PauseApps(ctx context.Context) {
	if e.err != nil {
		return
	}

	e.Progressf(ctx, "pausing")
	err := e.grp.Pause(ctx)
	if err != nil {
		e.err = fmt.Errorf("pause: %w", err)
		return
	}

	t := time.NewTicker(10 * time.Millisecond)
	defer t.Stop()
	for range t.C {
		s := e.grp.Status()
		var pausing, waiting int
		for _, node := range s.Nodes {
			for _, task := range node.Tasks {
				if task.Name == "pause" {
					pausing++
				}
				if task.Name == "resume-after" {
					waiting++
				}
			}
		}

		if pausing == 0 && waiting == len(s.Nodes) {
			break
		}
		if waiting == 0 {
			e.err = fmt.Errorf("pause: timed out waiting for nodes to pause")
			return
		}
	}
}

// DisableTriggers will disable all triggers in the new DB.
func (e *Execute) DisableNextDBTriggers(ctx context.Context) {
	if e.err != nil {
		return
	}

	swogrp.Progressf(ctx, "disabling triggers")

	var send pgx.Batch
	for _, table := range e.tables {
		send.Queue(fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER USER", table.QuotedName()))
	}

	e.err = e.nextDBConn.SendBatch(ctx, &send).Close()
	if e.err != nil {
		e.err = fmt.Errorf("disable triggers on next DB: %w", e.err)
	}
}

// EnableTriggers will re-enable triggers in the new DB.
func (e *Execute) enableTriggers(ctx context.Context) error {
	var send pgx.Batch

	for _, table := range e.tables {
		send.Queue(fmt.Sprintf("ALTER TABLE %s ENABLE TRIGGER USER", table.QuotedName()))
	}

	e.err = e.nextDBConn.SendBatch(ctx, &send).Close()
	if e.err != nil {
		return fmt.Errorf("enable triggers on next DB: %w", e.err)
	}
	return nil
}

// stopTheWorld grabs the exclusive advisory lock and then ensures the current state
// is set to in_progress.
func (e *Execute) stopTheWorld(ctx context.Context, srcTx pgx.Tx) error {
	e.Progressf(ctx, "stop-the-world")
	_, err := srcTx.Exec(ctx, fmt.Sprintf("select pg_advisory_xact_lock(%d)", lock.GlobalSwitchOver))
	if err != nil {
		return err
	}

	var stat string
	err = srcTx.QueryRow(ctx, "select current_state from switchover_state nowait").Scan(&stat)
	if err != nil {
		return err
	}
	switch stat {
	case "in_progress":
		return nil
	case "use_next_db":
		return swogrp.ErrDone
	case "idle":
		return errors.New("not in progress")
	default:
		if e.err == nil {
			return errors.New("unknown state: " + stat)
		}
		return e.err
	}
}

// FinalSync will attempt to lock and finalize the switchover.
func (e *Execute) FinalSync(ctx context.Context) {
	if e.err != nil {
		return
	}

	e.Progressf(ctx, "finalizing")

	// set timeouts before waiting on locks
	e.exec(ctx, e.mainDBConn, "set idle_in_transaction_session_timeout = 3000")
	e.exec(ctx, e.mainDBConn, "set lock_timeout = 3000")
	e.SyncLoop(ctx)
	if e.err != nil {
		return
	}

	srcTx, dstTx, err := e.syncTx(ctx, false)
	if err != nil {
		e.err = fmt.Errorf("final sync: %w", err)
		return
	}
	defer srcTx.Rollback(ctx)
	defer dstTx.Rollback(ctx)

	if err = e.stopTheWorld(ctx, srcTx); err != nil {
		e.err = fmt.Errorf("final sync: stop-the-world: %w", err)
		return
	}

	go e.Progressf(ctx, "last sync")
	_, err = e.syncChanges(ctx, srcTx, dstTx)
	if err != nil {
		e.err = fmt.Errorf("sync change log: %w", err)
		return
	}

	if err = e.syncSequences(ctx, srcTx, dstTx); err != nil {
		e.err = fmt.Errorf("sync sequences: %w", err)
		return
	}

	if err = dstTx.Commit(ctx); err != nil {
		e.err = fmt.Errorf("commit dst: %w", err)
		return
	}

	if err = e.enableTriggers(ctx); err != nil {
		return
	}

	_, err = srcTx.Exec(ctx, "update switchover_state set current_state = 'use_next_db' where current_state = 'in_progress'")
	if err != nil {
		e.err = fmt.Errorf("update switchover state: %w", err)
		return
	}

	err = srcTx.Commit(ctx)
	if err != nil {
		e.err = fmt.Errorf("commit src: %w", err)
		return
	}

	e.Progressf(ctx, "done")
}

func (e *Execute) syncTx(ctx context.Context, readOnly bool) (src, dst pgx.Tx, err error) {
	var srcOpts pgx.TxOptions
	if readOnly {
		srcOpts = pgx.TxOptions{
			AccessMode:     pgx.ReadOnly,
			IsoLevel:       pgx.Serializable,
			DeferrableMode: pgx.Deferrable,
		}
	}

	srcTx, err := e.mainDBConn.BeginTx(ctx, srcOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("begin src: %w", err)
	}

	dstTx, err := e.nextDBConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		srcTx.Rollback(ctx)
		return nil, nil, fmt.Errorf("begin dst: %w", err)
	}

	return srcTx, dstTx, nil
}

func (t Table) IDs(ids []string) interface{} {
	switch t.IDCol.Type {
	case "integer", "bigint":
		return sqlutil.IntArray(intIDs(ids))
	case "uuid":
		return sqlutil.UUIDArray(ids)
	}
	return sqlutil.StringArray(ids)
}

type syncData struct {
	t        Table
	toInsert []syncRow
	toUpdate []syncRow
	toDelete []string
}

type syncRow struct {
	table string
	id    string
	data  json.RawMessage
}
type pgxQueryer interface {
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryFunc(context.Context, string, []interface{}, []interface{}, func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error)
}
