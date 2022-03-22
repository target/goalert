package swo

import (
	"context"
	"errors"
	"fmt"

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
		m.Progressf(ctx, "enabling change log")
		err := EnableChangeLog(ctx, oldConn)
		if err != nil {
			return fmt.Errorf("enable change log: %w", err)
		}

		m.Progressf(ctx, "disabling triggers")
		err = DisableTriggers(ctx, newConn)
		if err != nil {
			return fmt.Errorf("disable triggers: %w", err)
		}

		m.Progressf(ctx, "performing initial sync")
		err = m.InitialSync(ctx, oldConn, newConn)
		if err != nil {
			return fmt.Errorf("initial sync: %w", err)
		}

		// sync in a loop until DB is up-to-date
		// err = m.LoopSync(ctx, oldConn, newConn)

		return errors.New("not implemented")
	})
}

// DisableTriggers will disable all triggers in the new DB.
func DisableTriggers(ctx context.Context, conn *pgx.Conn) error {
	tables, err := ScanTables(ctx, conn)
	if err != nil {
		return fmt.Errorf("scan tables: %w", err)
	}

	for _, table := range tables {
		_, err := conn.Exec(ctx, fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER USER", table.QuotedName()))
		if err != nil {
			return fmt.Errorf("%s: %w", table.Name, err)
		}
	}

	return nil
}

func LoopSync(ctx context.Context, oldConn, newConn *pgx.Conn) error {
	return nil
}

func FinalSync(ctx context.Context, oldConn, newConn *pgx.Conn) error {
	return nil
}

func syncChanges(ctx context.Context, oldConn, newConn pgx.Tx) (int, error) {
	return 0, nil
}
