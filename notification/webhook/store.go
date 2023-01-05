package webhook

import (
	"context"
	"database/sql"
	"fmt"

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
		AND value = $1`),
	}, p.Err
}

func (store *Store) FindOne(ctx context.Context, url string) (*Webhook, error) {
	err := validate.URL("webhookURL", url)
	if err != nil {
		fmt.Println("validate URL error: ", url)
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	row := store.findOne.QueryRowContext(ctx, url)

	var webhook Webhook

	err = row.Scan(&webhook.ID, &webhook.Name)
	if err != nil {
		return nil, err
	}

	return &webhook, err
}
