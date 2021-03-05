package app

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine"

	"github.com/pkg/errors"
)

func (app *App) initEngine(ctx context.Context) error {

	var regionIndex int
	err := app.db.QueryRowContext(ctx, `SELECT id FROM region_ids WHERE name = $1`, app.cfg.RegionName).Scan(&regionIndex)
	if errors.Is(err, sql.ErrNoRows) {
		// doesn't exist, try to create
		err = app.db.QueryRowContext(ctx, `
		WITH inserted AS (
			INSERT INTO region_ids (name) VALUES ($1) 
			ON CONFLICT DO NOTHING
			RETURNING id
		)
		SELECT id FROM region_ids WHERE name = $1
		UNION
		SELECT id FROM inserted
	`, app.cfg.RegionName).Scan(&regionIndex)
	}
	if err != nil {
		return errors.Wrap(err, "get region index")
	}

	app.Engine, err = engine.NewEngine(ctx, app.db, &engine.Config{
		AlertStore:          app.AlertStore,
		AlertLogStore:       app.AlertLogStore,
		ContactMethodStore:  app.ContactMethodStore,
		NotificationManager: app.notificationManager,
		UserStore:           app.UserStore,
		NotificationStore:   app.NotificationStore,
		NCStore:             app.NCStore,

		ConfigSource: app.ConfigStore,

		Keys: app.cfg.EncryptionKeys,

		MaxMessages: 50,

		DisableCycle: app.cfg.APIOnly,
	})
	if err != nil {
		return errors.Wrap(err, "init engine")
	}

	app.notificationManager.RegisterReceiver(app.Engine)

	return nil
}
