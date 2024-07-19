package engine

import (
	"context"
	"database/sql"

	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
)

type backend struct {
	db *sql.DB

	findOne *sql.Stmt

	trackStatus *sql.Stmt

	clientID string

	validCM *sql.Stmt
}

func newBackend(db *sql.DB) (*backend, error) {
	p := &util.Prepare{DB: db}

	return &backend{
		db:       db,
		clientID: uuid.New().String(),

		findOne: p.P(`
			SELECT
				id,
				alert_id,
				service_id,
				contact_method_id,
				created_at
			FROM outgoing_messages
			WHERE id = $1
		`),

		trackStatus: p.P(`
			insert into alert_status_subscriptions (channel_id, contact_method_id, alert_id, last_alert_status)
			values ($1, $2, $3, 'triggered')
		`),

		validCM: p.P(`select true from user_contact_methods where disabled = false and type = $1 and value = $2`),
	}, p.Err
}

func (b *backend) FindOne(ctx context.Context, id string) (*callback, error) {
	err := validate.UUID("CallbackID", id)
	if err != nil {
		return nil, err
	}

	var c callback
	var alertID sql.NullInt64
	var serviceID sql.NullString
	var cmID uuid.NullUUID
	err = b.findOne.QueryRowContext(ctx, id).Scan(&c.ID, &alertID, &serviceID, &cmID, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	c.AlertID = int(alertID.Int64)
	c.ServiceID = serviceID.String
	c.ContactMethodID = cmID.UUID
	return &c, nil
}
