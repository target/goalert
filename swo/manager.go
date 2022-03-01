package swo

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/swo/swomsg"
	"github.com/target/goalert/util/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Manager struct {
	id uuid.UUID

	dbOld, dbNew *gorm.DB
	protectedDB  *sql.DB

	s Syncer

	msgLog     *swomsg.Log
	nextMsgLog *swomsg.Log

	msgCh     chan *swomsg.Message
	nextMsgCh chan *swomsg.Message
	errCh     chan error

	nodes map[uuid.UUID]*Node
	exec  map[uuid.UUID]*swomsg.Message

	cancel func()

	canExec bool
}

type Node struct {
	ID uuid.UUID

	OldValid bool
	NewValid bool
}

func NewManager(dbcOld, dbcNew driver.Connector, canExec bool) (*Manager, error) {
	gCfg := &gorm.Config{PrepareStmt: true}
	gormOld, err := gorm.Open(postgres.New(postgres.Config{Conn: sql.OpenDB(dbcOld)}), gCfg)
	if err != nil {
		return nil, err
	}
	gormNew, err := gorm.Open(postgres.New(postgres.Config{Conn: sql.OpenDB(dbcNew)}), gCfg)
	if err != nil {
		return nil, err
	}

	id := uuid.New()
	msgLog, err := swomsg.NewLog(gormOld, id)
	if err != nil {
		return nil, err
	}

	msgLogNext, err := swomsg.NewLog(gormNew, id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		dbOld: gormOld,
		dbNew: gormNew,

		protectedDB: sql.OpenDB(NewConnector(dbcOld, dbcNew)),

		id:         id,
		msgLog:     msgLog,
		nextMsgLog: msgLogNext,
		canExec:    canExec,
		msgCh:      make(chan *swomsg.Message),
		nextMsgCh:  make(chan *swomsg.Message),
		errCh:      make(chan error, 10),
		cancel:     cancel,
		nodes:      make(map[uuid.UUID]*Node),
		exec:       make(map[uuid.UUID]*swomsg.Message),
	}

	go func() {
		for {
			msg, err := m.msgLog.Next(ctx)
			if err != nil {
				m.errCh <- fmt.Errorf("read from log: %w", err)
				return
			}
			m.msgCh <- msg
		}
	}()
	go func() {
		msg, err := m.nextMsgLog.Next(ctx)
		if err != nil {
			m.errCh <- fmt.Errorf("read from next log: %w", err)
			return
		}
		m.nextMsgCh <- msg
	}()

	go m.loop(ctx)

	return m, nil
}

func (m *Manager) DB() *sql.DB { return m.protectedDB }

func (m *Manager) processMessage(ctx context.Context, msg *swomsg.Message) {
	appendLog := func(msg interface{}) {
		err := m.msgLog.Append(ctx, msg)
		if err != nil {
			log.Log(ctx, err)
		}
	}

	switch {
	case msg.Ping != nil:
		appendLog(swomsg.Pong{IsNextDB: false})
		err := m.nextMsgLog.Append(ctx, swomsg.Pong{IsNextDB: true})
		if err != nil {
			log.Log(ctx, err)
		}
	case msg.Reset != nil:
		m.nodes = make(map[uuid.UUID]*Node)
		m.exec = make(map[uuid.UUID]*swomsg.Message)
		m.id = uuid.New()
	}

	if !m.canExec {
		// api-only node, don't process execute commands
		return
	}

	// any execute command needs to be claimed
	switch {
	case msg.Execute != nil:
		m.exec[msg.ID] = msg
		appendLog(swomsg.Claim{MsgID: msg.ID})
	case msg.Reset != nil:
		m.exec[msg.ID] = msg
		appendLog(swomsg.Claim{MsgID: msg.ID})
	case msg.Claim != nil:
		execMsg := m.exec[msg.Claim.MsgID]
		delete(m.exec, msg.Claim.MsgID)
		if msg.NodeID != m.id {
			// claimed by another node
			return
		}

		m.execute(execMsg)
	}
}

func (m *Manager) execute(msg *swomsg.Message) {
	switch {
	}
}

func (m *Manager) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-m.msgCh:
			m.processMessage(ctx, msg)
		case msg := <-m.nextMsgCh:
			if msg.Pong != nil && msg.Pong.IsNextDB {
				m.nodes[msg.NodeID].NewValid = true
			}
		case err := <-m.errCh:
			log.Log(ctx, err)
			m.msgLog.Append(ctx, swomsg.Error{Details: err.Error()})
		}
	}
}
