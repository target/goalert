package swo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/swo/swogrp"
	"github.com/target/goalert/swo/swoinfo"
	"github.com/target/goalert/swo/swomsg"
	"github.com/target/goalert/swo/swosync"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqldrv"
	"github.com/target/goalert/version"
)

// A Manager is responsible for managing the switchover process.
type Manager struct {
	// sql.DB instance safe for the application to use (instrumented for safe SWO operation)
	dbApp  *sql.DB
	dbMain *pgxpool.Pool
	dbNext *pgxpool.Pool

	pauseResume lifecycle.PauseResumer

	Config

	taskMgr *swogrp.TaskMgr

	MainDBInfo *swoinfo.DB
	NextDBInfo *swoinfo.DB
}

// Node contains information on a GoAlert instance in SWO mode.
type Node struct {
	ID uuid.UUID

	// OldValid indicates that the old database config is valid.
	OldValid bool

	// NewValid indicates that the new database config is valid.
	NewValid bool

	// CanExec indicates the node is NOT in API-only mode and is capable of executing tasks.
	CanExec bool

	Status string
}

// Config configures the current node for SWO.
type Config struct {
	OldDBURL, NewDBURL string
	CanExec            bool
	Logger             *log.Logger
}

// NewManager will create a new Manager with the given configuration.
func NewManager(cfg Config) (*Manager, error) {
	id := uuid.New()

	appStr := func(typ ConnType) string {
		return ConnInfo{
			Version: version.GitVersion(),
			ID:      id,
			Type:    typ,
		}.String()
	}

	mainURL, err := sqldrv.AppURL(cfg.OldDBURL, appStr(ConnTypeMainMgr))
	if err != nil {
		return nil, fmt.Errorf("connect to old db: %w", err)
	}
	mainPool, err := pgxpool.New(context.Background(), mainURL)
	if err != nil {
		return nil, fmt.Errorf("connect to old db: %w", err)
	}
	mainAppDBC, err := sqldrv.NewConnector(cfg.OldDBURL, appStr(ConnTypeMainApp))
	if err != nil {
		return nil, fmt.Errorf("connect to old db: %w", err)
	}
	nextURL, err := sqldrv.AppURL(cfg.NewDBURL, appStr(ConnTypeNextMgr))
	if err != nil {
		return nil, fmt.Errorf("connect to new db: %w", err)
	}
	nextPool, err := pgxpool.New(context.Background(), nextURL)
	if err != nil {
		return nil, fmt.Errorf("connect to new db: %w", err)
	}
	nextAppDBC, err := sqldrv.NewConnector(cfg.NewDBURL, appStr(ConnTypeNextApp))
	if err != nil {
		return nil, fmt.Errorf("connect to new db: %w", err)
	}

	m := &Manager{
		Config: cfg,
		dbApp:  sql.OpenDB(NewConnector(mainAppDBC, nextAppDBC)),
		dbMain: mainPool,
		dbNext: nextPool,
	}

	ctx := cfg.Logger.BackgroundContext()
	messages, err := swomsg.NewLog(ctx, m.dbMain)
	if err != nil {
		return nil, err
	}

	err = m.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
		var err error
		m.MainDBInfo, err = swoinfo.DBInfo(ctx, oldConn)
		if err != nil {
			return err
		}
		m.NextDBInfo, err = swoinfo.DBInfo(ctx, newConn)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("et server version: %w", err)
	}

	m.taskMgr, err = swogrp.NewTaskMgr(ctx, swogrp.Config{
		NodeID:  id,
		CanExec: cfg.CanExec,

		Logger:   cfg.Logger,
		Messages: messages,

		OldID: m.MainDBInfo.ID,
		NewID: m.NextDBInfo.ID,

		Executor:   NewExecutor(m),
		PauseFunc:  func(ctx context.Context) error { return m.pauseResume.Pause(ctx) },
		ResumeFunc: func(ctx context.Context) error { return m.pauseResume.Resume(ctx) },
	})
	if err != nil {
		return nil, fmt.Errorf("init task manager: %w", err)
	}

	return m, nil
}

// SetPauseResumer allows setting the pause/resume functionality for the manager.
//
// Pause is called during the switchover process to minimize the number of
// long-lived DB connections so that the final sync can be performed quickly.
//
// After a switchover, or if it is aborted, Resume will be called.
func (m *Manager) SetPauseResumer(app lifecycle.PauseResumer) {
	if m.pauseResume != nil {
		panic("already set")
	}
	m.pauseResume = app
	m.taskMgr.Init()
}

// ConnInfo returns information about all current DB connections.
func (m *Manager) ConnInfo(ctx context.Context) (counts []swoinfo.ConnCount, err error) {
	err = m.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
		counts, err = swoinfo.ConnInfo(ctx, oldConn, newConn)
		return err
	})

	return
}

// withConnFromOld allows performing operations with a raw connection to the old database.
func (m *Manager) withConnFromOld(ctx context.Context, f func(context.Context, *pgx.Conn) error) error {
	return withPGXConn(ctx, m.dbMain, f)
}

// withConnFromBoth allows performing operations with a raw connection to both databases database.
func (m *Manager) withConnFromBoth(ctx context.Context, f func(ctx context.Context, oldConn, newConn *pgx.Conn) error) error {
	// grab lock with old DB first
	return withPGXConn(ctx, m.dbMain, func(ctx context.Context, connMain *pgx.Conn) error {
		return withPGXConn(ctx, m.dbNext, func(ctx context.Context, connNext *pgx.Conn) error {
			return f(ctx, connMain, connNext)
		})
	})
}

func withPGXConn(ctx context.Context, db *pgxpool.Pool, runFunc func(context.Context, *pgx.Conn) error) error {
	return db.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
		return runFunc(ctx, conn.Conn())
	})
}

// Status will return the current switchover status.
func (m *Manager) Status() Status {
	return Status{
		Status:        m.taskMgr.Status(),
		MainDBID:      m.MainDBInfo.ID,
		NextDBID:      m.NextDBInfo.ID,
		MainDBVersion: m.MainDBInfo.Version,
		NextDBVersion: m.NextDBInfo.Version,
	}
}

// Reset will disable the changelog and reset the cluster state.
func (m *Manager) Reset(ctx context.Context) error {
	err := m.taskMgr.Cancel(ctx)
	if err != nil {
		return fmt.Errorf("cancel task: %w", err)
	}

	err = m.withConnFromOld(ctx, func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, swosync.ConnWaitLockQuery)
		if err != nil {
			return err
		}

		return swodb.New(conn).DisableChangeLogTriggers(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to disable change log triggers: %w", err)
	}

	err = m.taskMgr.Reset(ctx)
	if err != nil {
		return fmt.Errorf("reset cluster state: %w", err)
	}

	return nil
}

// StartExecute will trigger the switchover to begin.
func (m *Manager) StartExecute(ctx context.Context) error { return m.taskMgr.Execute(ctx) }

// DB returns a sql.DB that will always return safe connections to be used during the switchover.
//
// All application code/queries should use this DB.
func (m *Manager) DB() *sql.DB { return m.dbApp }
