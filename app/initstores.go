package app

import (
	"context"
	"net/url"

	"github.com/target/goalert/calendarsubscription"

	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/auth/nonce"
	"github.com/target/goalert/config"
	"github.com/target/goalert/engine/resolver"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/label"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/override"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/service"
	"github.com/target/goalert/timezone"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/favorite"
	"github.com/target/goalert/user/notificationrule"

	"github.com/pkg/errors"
)

func (app *App) initStores(ctx context.Context) error {
	var err error

	if app.ConfigStore == nil {
		var fallback url.URL
		fallback.Scheme = "http"
		fallback.Host = app.l.Addr().String()
		app.ConfigStore, err = config.NewStore(ctx, app.db, app.cfg.EncryptionKeys, fallback.String())
	}
	if err != nil {
		return errors.Wrap(err, "init config store")
	}

	if app.NonceStore == nil {
		app.NonceStore, err = nonce.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init nonce store")
	}

	if app.OAuthKeyring == nil {
		app.OAuthKeyring, err = keyring.NewDB(ctx, app.db, &keyring.Config{
			Name:         "oauth-state",
			RotationDays: 1,
			MaxOldKeys:   1,
			Keys:         app.cfg.EncryptionKeys,
		})
	}
	if err != nil {
		return errors.Wrap(err, "init oauth state keyring")
	}

	if app.SessionKeyring == nil {
		app.SessionKeyring, err = keyring.NewDB(ctx, app.db, &keyring.Config{
			Name:         "browser-sessions",
			RotationDays: 1,
			MaxOldKeys:   30,
			Keys:         app.cfg.EncryptionKeys,
		})
	}
	if err != nil {
		return errors.Wrap(err, "init session keyring")
	}

	if app.AlertLogStore == nil {
		app.AlertLogStore, err = alertlog.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init alertlog store")
	}

	if app.AlertStore == nil {
		app.AlertStore, err = alert.NewDB(ctx, app.db, app.AlertLogStore)
	}
	if err != nil {
		return errors.Wrap(err, "init alert store")
	}

	if app.ContactMethodStore == nil {
		app.ContactMethodStore, err = contactmethod.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init contact method store")
	}

	if app.NotificationRuleStore == nil {
		app.NotificationRuleStore, err = notificationrule.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init notification rule store")
	}

	if app.ServiceStore == nil {
		app.ServiceStore, err = service.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init service store")
	}
	if app.ScheduleStore == nil {
		app.ScheduleStore, err = schedule.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init schedule store")
	}

	if app.CalendarSubscriptionStore == nil {
		app.CalendarSubscriptionStore, err = calendarsubscription.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init calendar subscription store")
	}
	if app.RotationStore == nil {
		app.RotationStore, err = rotation.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init rotation store")
	}

	if app.UserStore == nil {
		app.UserStore, err = user.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init user store")
	}

	if app.NCStore == nil {
		app.NCStore, err = notificationchannel.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init notification channel store")
	}

	if app.EscalationStore == nil {
		app.EscalationStore, err = escalation.NewDB(ctx, app.db, escalation.Config{
			LogStore: app.AlertLogStore,
			NCStore:  app.NCStore,
			SlackLookupFunc: func(ctx context.Context, channelID string) (*slack.Channel, error) {
				return app.slackChan.Channel(ctx, channelID)
			},
		})
	}
	if err != nil {
		return errors.Wrap(err, "init escalation policy store")
	}

	if app.IntegrationKeyStore == nil {
		app.IntegrationKeyStore, err = integrationkey.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init integration key store")
	}

	if app.ScheduleRuleStore == nil {
		app.ScheduleRuleStore, err = rule.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init schedule rule store")
	}

	if app.NotificationStore == nil {
		app.NotificationStore, err = notification.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init notification store")
	}

	if app.FavoriteStore == nil {
		app.FavoriteStore, err = favorite.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init favorite store")
	}

	if app.OverrideStore == nil {
		app.OverrideStore, err = override.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init override store")
	}

	if app.Resolver == nil {
		app.Resolver, err = resolver.NewDB(ctx, app.db, app.ScheduleRuleStore, app.ScheduleStore)
	}
	if err != nil {
		return errors.Wrap(err, "init resolver")
	}

	if app.LimitStore == nil {
		app.LimitStore, err = limit.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init limit config store")
	}
	if app.HeartbeatStore == nil {
		app.HeartbeatStore, err = heartbeat.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init heartbeat store")
	}
	if app.LabelStore == nil {
		app.LabelStore, err = label.NewDB(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init label store")
	}

	if app.OnCallStore == nil {
		app.OnCallStore, err = oncall.NewDB(ctx, app.db, app.ScheduleRuleStore)
	}
	if err != nil {
		return errors.Wrap(err, "init on-call store")
	}

	if app.TimeZoneStore == nil {
		app.TimeZoneStore = timezone.NewStore(ctx, app.db)
	}

	return nil
}
