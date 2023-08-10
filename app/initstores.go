package app

import (
	"context"
	"net/url"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/alert/alertmetrics"
	"github.com/target/goalert/auth/authlink"
	"github.com/target/goalert/auth/basic"
	"github.com/target/goalert/auth/nonce"
	"github.com/target/goalert/calsub"
	"github.com/target/goalert/config"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/label"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/notice"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
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
		fallback.Path = app.cfg.HTTPPrefix
		storeCfg := config.StoreConfig{
			DB:                 app.db,
			Keys:               app.cfg.EncryptionKeys,
			FallbackURL:        fallback.String(),
			ExplicitURL:        app.cfg.PublicURL,
			IngressEmailDomain: app.cfg.EmailIntegrationDomain,
		}
		app.ConfigStore, err = config.NewStore(ctx, storeCfg)
	}
	if err != nil {
		return errors.Wrap(err, "init config store")
	}
	if app.cfg.InitialConfig != nil {
		permission.SudoContext(ctx, func(ctx context.Context) {
			err = app.ConfigStore.SetConfig(ctx, *app.cfg.InitialConfig)
		})
		if err != nil {
			return errors.Wrap(err, "set initial config")
		}
	}

	if app.NonceStore == nil {
		app.NonceStore, err = nonce.NewStore(ctx, app.cfg.Logger, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init nonce store")
	}

	if app.OAuthKeyring == nil {
		app.OAuthKeyring, err = keyring.NewDB(ctx, app.cfg.Logger, app.db, &keyring.Config{
			Name:         "oauth-state",
			RotationDays: 1,
			MaxOldKeys:   1,
			Keys:         app.cfg.EncryptionKeys,
		})
	}
	if err != nil {
		return errors.Wrap(err, "init oauth state keyring")
	}

	if app.AuthLinkKeyring == nil {
		app.AuthLinkKeyring, err = keyring.NewDB(ctx, app.cfg.Logger, app.db, &keyring.Config{
			Name:         "auth-link",
			RotationDays: 1,
			MaxOldKeys:   1,
			Keys:         app.cfg.EncryptionKeys,
		})
	}
	if err != nil {
		return errors.Wrap(err, "init oauth state keyring")
	}

	if app.SessionKeyring == nil {
		app.SessionKeyring, err = keyring.NewDB(ctx, app.cfg.Logger, app.db, &keyring.Config{
			Name:         "browser-sessions",
			RotationDays: 1,
			MaxOldKeys:   30,
			Keys:         app.cfg.EncryptionKeys,
		})
	}
	if err != nil {
		return errors.Wrap(err, "init session keyring")
	}

	if app.APIKeyring == nil {
		app.APIKeyring, err = keyring.NewDB(ctx, app.cfg.Logger, app.db, &keyring.Config{
			Name:       "api-keys",
			MaxOldKeys: 100,
			Keys:       app.cfg.EncryptionKeys,
		})
	}
	if err != nil {
		return errors.Wrap(err, "init API keyring")
	}

	if app.AuthLinkStore == nil {
		app.AuthLinkStore, err = authlink.NewStore(ctx, app.db, app.AuthLinkKeyring)
	}
	if err != nil {
		return errors.Wrap(err, "init auth link store")
	}

	if app.AlertMetricsStore == nil {
		app.AlertMetricsStore, err = alertmetrics.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init alert metrics store")
	}

	if app.AlertLogStore == nil {
		app.AlertLogStore, err = alertlog.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init alertlog store")
	}

	if app.AlertStore == nil {
		app.AlertStore, err = alert.NewStore(ctx, app.db, app.AlertLogStore)
	}
	if err != nil {
		return errors.Wrap(err, "init alert store")
	}

	if app.ContactMethodStore == nil {
		app.ContactMethodStore, err = contactmethod.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init contact method store")
	}

	if app.NotificationRuleStore == nil {
		app.NotificationRuleStore, err = notificationrule.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init notification rule store")
	}

	if app.ServiceStore == nil {
		app.ServiceStore, err = service.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init service store")
	}

	if app.AuthBasicStore == nil {
		app.AuthBasicStore, err = basic.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init basic auth store")
	}

	if app.UserStore == nil {
		app.UserStore, err = user.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init user store")
	}

	if app.ScheduleStore == nil {
		app.ScheduleStore, err = schedule.NewStore(ctx, app.db, app.UserStore)
	}
	if err != nil {
		return errors.Wrap(err, "init schedule store")
	}

	if app.RotationStore == nil {
		app.RotationStore, err = rotation.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init rotation store")
	}

	if app.NCStore == nil {
		app.NCStore, err = notificationchannel.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init notification channel store")
	}

	if app.EscalationStore == nil {
		app.EscalationStore, err = escalation.NewStore(ctx, app.db, escalation.Config{
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
		app.IntegrationKeyStore, err = integrationkey.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init integration key store")
	}

	if app.ScheduleRuleStore == nil {
		app.ScheduleRuleStore, err = rule.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init schedule rule store")
	}

	if app.NotificationStore == nil {
		app.NotificationStore, err = notification.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init notification store")
	}

	if app.FavoriteStore == nil {
		app.FavoriteStore, err = favorite.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init favorite store")
	}

	if app.OverrideStore == nil {
		app.OverrideStore, err = override.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init override store")
	}

	if app.LimitStore == nil {
		app.LimitStore, err = limit.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init limit config store")
	}
	if app.HeartbeatStore == nil {
		app.HeartbeatStore, err = heartbeat.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init heartbeat store")
	}
	if app.LabelStore == nil {
		app.LabelStore, err = label.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init label store")
	}

	if app.OnCallStore == nil {
		app.OnCallStore, err = oncall.NewStore(ctx, app.db, app.ScheduleRuleStore, app.ScheduleStore)
	}
	if err != nil {
		return errors.Wrap(err, "init on-call store")
	}

	if app.TimeZoneStore == nil {
		app.TimeZoneStore = timezone.NewStore(ctx, app.db)
	}

	if app.CalSubStore == nil {
		app.CalSubStore, err = calsub.NewStore(ctx, app.db, app.APIKeyring, app.OnCallStore)
	}
	if err != nil {
		return errors.Wrap(err, "init calendar subscription store")
	}

	if app.NoticeStore == nil {
		app.NoticeStore, err = notice.NewStore(ctx, app.db)
	}
	if err != nil {
		return errors.Wrap(err, "init notice store")
	}

	return nil
}
