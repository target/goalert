package dbsync

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	"github.com/pkg/errors"
	"github.com/target/goalert/lock"
	"github.com/target/goalert/switchover"
	"github.com/vbauerster/mpb/v4"
)

type Sync struct {
	oldDB, newDB *sql.DB
	newURL       string
	oldOffset    time.Duration
	newOffset    time.Duration
	tables       []Table
	nodeStatus   map[string]switchover.Status
	mx           sync.Mutex
	statChange   chan struct{}

	oldDBID, newDBID string
}

func (s *Sync) RefreshTables(ctx context.Context) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	t, err := Tables(ctx, s.oldDB)
	if err != nil {
		return err
	}
	s.tables = t
	return nil
}
func NewSync(ctx context.Context, oldDB, newDB *sql.DB, newURL string) (*Sync, error) {
	oldOffset, err := switchover.CalcDBOffset(ctx, oldDB)
	if err != nil {
		return nil, err
	}

	newOffset, err := switchover.CalcDBOffset(ctx, newDB)
	if err != nil {
		return nil, err
	}

	s := &Sync{
		oldDB:      oldDB,
		newDB:      newDB,
		oldDBID:    newDBID(),
		newDBID:    newDBID(),
		newURL:     newURL,
		oldOffset:  oldOffset,
		newOffset:  newOffset,
		nodeStatus: make(map[string]switchover.Status),
		statChange: make(chan struct{}),
	}

	err = s.RefreshTables(ctx)
	if err != nil {
		return nil, err
	}

	go s.listen(s.oldDB)
	go s.listen(s.newDB)

	return s, nil
}
func (s *Sync) Offset() time.Duration {
	return s.oldOffset
}
func (s *Sync) NodeStatus() []switchover.Status {
	var stat []switchover.Status
	s.mx.Lock()
	for _, st := range s.nodeStatus {
		stat = append(stat, st)
	}
	s.mx.Unlock()
	sort.Slice(stat, func(i, j int) bool { return stat[i].NodeID < stat[j].NodeID })
	return stat
}

type WaitState struct {
	Done  int
	Total int
	Abort bool
}

func (s *Sync) NodeStateWait(ctx context.Context, origTotal int, bar *mpb.Bar, anyState ...switchover.State) error {
	for {
		abort, n, total := s.nodeStateAll(anyState...)
		if abort {
			return errors.New("node abort")
		}
		if total != origTotal {
			return errors.New("new node appeared while waiting")
		}
		cur := bar.Current()
		if n != cur {
			bar.IncrBy(int(n-cur), time.Second)
		}
		if int(n) == total {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.statChange:
			continue
		}
	}
}
func (s *Sync) Ready() bool {
	a, c, n := s.nodeStateAll(switchover.StateReady)
	if a {
		return false
	}
	if int(c) != n {
		return false
	}
	if n == 0 {
		return false
	}
	return true
}
func (s *Sync) nodeStateAll(anyState ...switchover.State) (bool, int64, int) {
	s.mx.Lock()
	defer s.mx.Unlock()
	var count int64
nodeCheck:
	for _, stat := range s.nodeStatus {
		if stat.State == switchover.StateAbort {
			return true, 0, 0
		}
		for _, state := range anyState {
			if state == stat.State {
				count++
				continue nodeCheck
			}
		}
	}
	return false, count, len(s.nodeStatus)
}
func (s *Sync) Aborted() bool {
	s.mx.Lock()
	defer s.mx.Unlock()
	for _, stat := range s.nodeStatus {
		if stat.State == switchover.StateAbort {
			return true
		}
	}
	return false
}

type progWrite struct {
	inc1 func(int, ...time.Duration)
	inc2 func(int, ...time.Duration)
}

func (w *progWrite) Write(p []byte) (int, error) {
	n := bytes.Count(p, []byte{'\n'})
	w.inc1(n)
	w.inc2(n)
	return len(p), nil
}

func (s *Sync) table(tableName string) Table {
	for _, t := range s.tables {
		if t.Name != tableName {
			continue
		}
		return t
	}
	panic("unknown table: " + tableName)
}

func (s *Sync) Sync(ctx context.Context, isFinal, enableSwitchOver bool) error {
	var stat string

	srcConn, err := stdlib.AcquireConn(s.oldDB)
	if err != nil {
		return errors.Wrap(err, "get src conn")
	}
	defer stdlib.ReleaseConn(s.oldDB, srcConn)
	defer srcConn.Close()

	var gotLock bool
	if isFinal {
		d, ok := ctx.Deadline()
		if !ok {
			return errors.New("context missing deadline for final sync")
		}

		lockMs := int64(time.Until(d.Add(-time.Second)) / time.Millisecond)
		if lockMs < 0 {
			return errors.New("not enough time remaining for lock")
		}

		_, err = srcConn.ExecEx(ctx, `set lock_timeout = `+strconv.FormatInt(lockMs, 10), nil)
		if err != nil {
			return errors.Wrap(err, "set lock_timeout")
		}
		_, err = srcConn.ExecEx(ctx, `select pg_advisory_lock($1)`, nil, lock.GlobalSwitchOver)
		if err == nil {
			gotLock = true
		}
	} else {
		err = srcConn.QueryRowEx(ctx, `select pg_try_advisory_lock_shared($1)`, nil, lock.GlobalSwitchOver).Scan(&gotLock)
	}
	if err != nil {
		return errors.Wrap(err, "acquire advisory lock")
	}
	if !gotLock {
		return errors.New("failed to get lock")
	}

	err = srcConn.QueryRowEx(ctx, `select current_state from switchover_state nowait`, nil).Scan(&stat)
	if err != nil {
		return errors.Wrap(err, "get current state")
	}
	if stat == "use_next_db" {
		return errors.New("switchover already completed")
	}
	if stat == "idle" {
		return errors.New("run enable first")
	}

	dstConn, err := stdlib.AcquireConn(s.newDB)
	if err != nil {
		return errors.Wrap(err, "get dst conn")
	}
	defer stdlib.ReleaseConn(s.newDB, dstConn)

	start := time.Now()
	txSrc, err := srcConn.BeginEx(ctx, &pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadOnly,
		DeferrableMode: pgx.Deferrable,
	})
	if err != nil {
		return errors.Wrap(err, "start src transaction")
	}
	defer txSrc.Rollback()
	fmt.Println("Got src tx after", time.Since(start))

	txDst, err := dstConn.BeginEx(ctx, &pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.Deferrable,
	})
	if err != nil {
		return errors.Wrap(err, "start dst transaction")
	}
	defer txDst.Rollback()

	_, err = txDst.ExecEx(ctx, `SET CONSTRAINTS ALL DEFERRED`, nil)
	if err != nil {
		return errors.Wrap(err, "defer constraints")
	}

	var srcLastChange, dstLastChange int
	err = txSrc.QueryRowEx(ctx, `select coalesce(max(id), 0) from change_log`, nil).Scan(&srcLastChange)
	if err != nil {
		return errors.Wrap(err, "check src last change")
	}
	err = txDst.QueryRowEx(ctx, `select coalesce(max(id), 0) from change_log`, nil).Scan(&dstLastChange)
	if err != nil {
		return errors.Wrap(err, "check dst last change")
	}
	if srcLastChange == 0 {
		return errors.New("change_log  (or no changes) on src DB")
	}

	if !isFinal {
		start = time.Now()
		batch := txDst.BeginBatch()
		for _, t := range s.tables {
			batch.Queue(`alter table `+t.SafeName()+` disable trigger user`, nil, nil, nil)
		}
		err = batch.Send(ctx, nil)
		if err != nil {
			return errors.Wrap(err, "disable triggers (send)")
		}
		err = batch.Close()
		if err != nil {
			return errors.Wrap(err, "disable triggers")
		}
		fmt.Println("Disabled destination triggers in", time.Since(start))
	}

	if dstLastChange == 0 {
		err = s.initialSync(ctx, txSrc, txDst)
	} else if srcLastChange > dstLastChange {
		err = s.diffSync(ctx, txSrc, txDst, dstLastChange)
	}
	if err != nil {
		return errors.Wrap(err, "sync")
	}

	start = time.Now()
	err = s.syncSequences(ctx, txSrc, txDst)
	if err != nil {
		return errors.Wrap(err, "update sequence numbers")
	}
	fmt.Println("Updated sequences in", time.Since(start))

	err = txSrc.Commit()
	if err != nil {
		return errors.Wrap(err, "commit src")
	}

	err = txDst.Commit()
	if err != nil {
		return errors.Wrap(err, "commit dst")
	}

	if isFinal {
		start = time.Now()
		batch := dstConn.BeginBatch()
		for _, t := range s.tables {
			batch.Queue(`alter table `+t.SafeName()+` enable trigger user`, nil, nil, nil)
		}
		err = batch.Send(ctx, nil)
		if err != nil {
			return errors.Wrap(err, "enable triggers (send)")
		}
		err = batch.Close()
		if err != nil {
			return errors.Wrap(err, "enable triggers")
		}
		fmt.Println("Re-enabled triggers in", time.Since(start))

		if enableSwitchOver {
			_, err = srcConn.ExecEx(ctx, `update switchover_state set current_state = 'use_next_db'`, nil)
			if err != nil {
				return errors.Wrap(err, "update state table")
			}
			fmt.Println("State updated: next-db is now active!")
		}
	}

	return nil
}
