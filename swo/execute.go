package swo

import (
	"context"
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

func (m *Manager) Execute(ctx context.Context, oldConn, newConn *pgx.Conn) error {
	/*
		- initial sync
		- loop until few changes
		- send proposal
		- loop until proposal is valid
		- execute proposal

	*/

	return m.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
		err := EnableChangeLog(ctx, oldConn)
		if err != nil {
			return fmt.Errorf("enable change log: %w", err)
		}

		err = m.InitialSync(ctx, oldConn, newConn)
		if err != nil {
			return fmt.Errorf("initial sync: %w", err)
		}

		// sync in a loop until DB is up-to-date
		// err = m.LoopSync(ctx, oldConn, newConn)

		return nil
	})
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
