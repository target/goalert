package app

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine"
	"github.com/target/goalert/notification"

	"github.com/pkg/errors"
)

func (app *App) initEngine(ctx context.Context) error {
	var regionIndex int
	err := app.db.QueryRowContext(ctx, `SELECT id FROM region_ids WHERE name = $1`, app.cfg.RegionName).Scan(&regionIndex)
	if errors.Is(err, sql.ErrNoRows) {
		// doesn't exist, try to create
		_, err = app.db.ExecContext(ctx, `
		INSERT INTO region_ids (name) VALUES ($1) 
		ON CONFLICT DO NOTHING`, app.cfg.RegionName)
		if err != nil {
			return errors.Wrap(err, "insert region")
		}

		err = app.db.QueryRowContext(ctx, `SELECT id FROM region_ids WHERE name = $1`, app.cfg.RegionName).Scan(&regionIndex)
	}
	if err != nil {
		return errors.Wrap(err, "get region index")
	}
	app.notificationManager = notification.NewManager(app.DestRegistry)
	app.Engine, err = engine.NewEngine(ctx, app.db, &engine.Config{
		AlertStore:          app.AlertStore,
		AlertLogStore:       app.AlertLogStore,
		ContactMethodStore:  app.ContactMethodStore,
		NotificationManager: app.notificationManager,
		UserStore:           app.UserStore,
		NotificationStore:   app.NotificationStore,
		NCStore:             app.NCStore,
		OnCallStore:         app.OnCallStore,
		ScheduleStore:       app.ScheduleStore,
		AuthLinkStore:       app.AuthLinkStore,
		SlackStore:          app.slackChan,
		DestRegistry:        app.DestRegistry,

		ConfigSource: app.ConfigStore,

		Keys: app.cfg.EncryptionKeys,

		CycleTime: app.cfg.EngineCycleTime,

		MaxMessages: 50,

		DisableCycle: app.cfg.APIOnly,
		LogCycles:    app.cfg.LogEngine,
		River:        app.River,
		RiverWorkers: app.RiverWorkers,
	})
	if err != nil {
		return errors.Wrap(err, "init engine")
	}

	return nil
}
