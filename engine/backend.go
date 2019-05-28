package engine

import (
	"context"
	"database/sql"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

type backend struct {
	db *sql.DB

	findOne *sql.Stmt

	clientID string
}

func newBackend(db *sql.DB) (*backend, error) {
	p := &util.Prepare{DB: db}

	return &backend{
		db:       db,
		clientID: uuid.NewV4().String(),

		findOne: p.P(`
			SELECT
				id,
				alert_id,
				contact_method_id
			FROM outgoing_messages
			WHERE id = $1
		`),
	}, p.Err
}

func (b *backend) FindOne(ctx context.Context, id string) (*callback, error) {
	err := validate.UUID("CallbackID", id)
	if err != nil {
		return nil, err
	}
	var c callback
	err = b.findOne.QueryRowContext(ctx, id).Scan(c.fields()...)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
