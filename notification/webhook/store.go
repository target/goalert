package webhook

import (
	"context"
	"database/sql"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

type Webhook struct {
	ID   string
	Name string
}

type Store struct {
	db *sql.DB

	findOne *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,
		findOne: p.P(`
		SELECT
			value,
			name
		FROM notification_channels
		WHERE type = 'WEBHOOK'
		AND id = $1`),
	}, p.Err
}

func (store *Store) FindOne(ctx context.Context, id string) (*Webhook, error) {
	err := validate.UUID("webhook_id", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	row := store.findOne.QueryRowContext(ctx, id)

	var webhook Webhook

	err = row.Scan(&webhook.ID, &webhook.Name)
	if err != nil {
		return nil, err
	}

	return &webhook, err
}
