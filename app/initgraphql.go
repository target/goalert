package app

import (
	"context"

	"github.com/target/goalert/graphql"
	"github.com/target/goalert/graphql2/graphqlapp"
	"github.com/target/goalert/schedule/shiftcalc"
)

func (app *App) initGraphQL(ctx context.Context) error {

	shiftC := &shiftcalc.ShiftCalculator{
		RotStore:   app.RotationStore,
		RuleStore:  app.ScheduleRuleStore,
		SchedStore: app.ScheduleStore,
		OStore:     app.OverrideStore,
	}

	app.graphql2 = &graphqlapp.App{
		DB:                app.db,
		UserStore:         app.UserStore,
		CMStore:           app.ContactMethodStore,
		NRStore:           app.NotificationRuleStore,
		NCStore:           app.NCStore,
		AlertStore:        app.AlertStore,
		AlertLogStore:     app.AlertLogStore,
		ServiceStore:      app.ServiceStore,
		FavoriteStore:     app.FavoriteStore,
		PolicyStore:       app.EscalationStore,
		ScheduleStore:     app.ScheduleStore,
		CalSubStore:       app.CalSubStore,
		RotationStore:     app.RotationStore,
		OnCallStore:       app.OnCallStore,
		TimeZoneStore:     app.TimeZoneStore,
		IntKeyStore:       app.IntegrationKeyStore,
		LabelStore:        app.LabelStore,
		RuleStore:         app.ScheduleRuleStore,
		OverrideStore:     app.OverrideStore,
		ConfigStore:       app.ConfigStore,
		LimitStore:        app.LimitStore,
		NotificationStore: app.NotificationStore,
		SlackStore:        app.slackChan,
		HeartbeatStore:    app.HeartbeatStore,
		NoticeStore:       *app.NoticeStore,
		Twilio:            app.twilioConfig,
	}

	var err error
	app.graphql, err = graphql.NewHandler(ctx, graphql.Config{
		DB:                  app.db,
		UserStore:           app.UserStore,
		AlertStore:          app.AlertStore,
		AlertLogStore:       app.AlertLogStore,
		CMStore:             app.ContactMethodStore,
		NRStore:             app.NotificationRuleStore,
		UserFavoriteStore:   app.FavoriteStore,
		ServiceStore:        app.ServiceStore,
		ScheduleStore:       app.ScheduleStore,
		RotationStore:       app.RotationStore,
		ShiftCalc:           shiftC,
		ScheduleRuleStore:   app.ScheduleRuleStore,
		EscalationStore:     app.EscalationStore,
		IntegrationKeyStore: app.IntegrationKeyStore,
		Resolver:            app.Resolver,
		NotificationStore:   app.NotificationStore,
		HeartbeatStore:      app.HeartbeatStore,
		OverrideStore:       app.OverrideStore,
		LabelStore:          app.LabelStore,
		OnCallStore:         app.OnCallStore,
	})

	return err
}
