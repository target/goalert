package app

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/riverqueue/river"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/alert/alertmetrics"
	"github.com/target/goalert/apikey"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/auth"
	"github.com/target/goalert/auth/authlink"
	"github.com/target/goalert/auth/basic"
	"github.com/target/goalert/auth/nonce"
	"github.com/target/goalert/calsub"
	"github.com/target/goalert/config"
	"github.com/target/goalert/engine"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/graphql2/graphqlapp"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/integrationkey/uik"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/label"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/notice"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/service"
	"github.com/target/goalert/smtpsrv"
	"github.com/target/goalert/timezone"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/favorite"
	"github.com/target/goalert/user/notificationrule"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
)

// App represents an instance of the GoAlert application.
type App struct {
	cfg Config

	Logger *slog.Logger

	mgr *lifecycle.Manager

	db     *sql.DB
	pgx    *pgxpool.Pool
	l      net.Listener
	events *sqlutil.Listener

	doneCh chan struct{}

	sysAPIL   net.Listener
	sysAPISrv *grpc.Server
	hSrv      *health.Server

	srv        *http.Server
	smtpsrv    *smtpsrv.Server
	smtpsrvL   net.Listener
	startupErr error

	notificationManager *notification.Manager
	Engine              *engine.Engine
	graphql2            *graphqlapp.App
	AuthHandler         *auth.Handler

	twilioSMS    *twilio.SMS
	twilioVoice  *twilio.Voice
	twilioConfig *twilio.Config

	slackChan *slack.ChannelSender

	ConfigStore *config.Store

	AlertStore        *alert.Store
	AlertLogStore     *alertlog.Store
	AlertMetricsStore *alertmetrics.Store

	AuthBasicStore        *basic.Store
	UserStore             *user.Store
	ContactMethodStore    *contactmethod.Store
	NotificationRuleStore *notificationrule.Store
	FavoriteStore         *favorite.Store

	ServiceStore        *service.Store
	EscalationStore     *escalation.Store
	IntegrationKeyStore *integrationkey.Store
	UIKHandler          *uik.Handler
	ScheduleRuleStore   *rule.Store
	NotificationStore   *notification.Store
	ScheduleStore       *schedule.Store
	RotationStore       *rotation.Store
	DestRegistry        *nfydest.Registry

	CalSubStore    *calsub.Store
	OverrideStore  *override.Store
	LimitStore     *limit.Store
	HeartbeatStore *heartbeat.Store

	OAuthKeyring    keyring.Keyring
	SessionKeyring  keyring.Keyring
	APIKeyring      keyring.Keyring
	AuthLinkKeyring keyring.Keyring

	NonceStore    *nonce.Store
	LabelStore    *label.Store
	OnCallStore   *oncall.Store
	NCStore       *notificationchannel.Store
	TimeZoneStore *timezone.Store
	NoticeStore   *notice.Store
	AuthLinkStore *authlink.Store
	APIKeyStore   *apikey.Store
	River   *river.Client[pgx.Tx]
	RiverUI *riverui.Server
}

// NewApp constructs a new App and binds the listening socket.
func NewApp(c Config, pool *pgxpool.Pool) (*App, error) {
	if c.Logger == nil {
		return nil, errors.New("Logger is required")
	}

	var err error
	db := stdlib.OpenDBFromPool(pool)
	permission.SudoContext(context.Background(), func(ctx context.Context) {
		// Should not be possible for the app to ever see `use_next_db` unless misconfigured.
		//
		// In switchover mode, the connector wrapper will check this and provide the app with
		// a connection to the next DB instead, if this was set.
		//
		// This is a sanity check to ensure that the app is not accidentally using the previous DB
		// after a switchover.
		err = db.QueryRowContext(ctx, `select true from switchover_state where current_state = 'use_next_db'`).Scan(new(bool))
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
			return
		}
		if err != nil {
			return
		}

		err = fmt.Errorf("refusing to connect to stale database (switchover_state table has use_next_db set)")
	})
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", c.ListenAddr)
	if err != nil {
		return nil, errors.Wrapf(err, "bind address %s", c.ListenAddr)
	}

	if c.TLSListenAddr != "" {
		l2, err := tls.Listen("tcp", c.TLSListenAddr, c.TLSConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "listen %s", c.TLSListenAddr)
		}
		l = newMultiListener(l, l2)
	}

	c.LegacyLogger.AddErrorMapper(func(ctx context.Context, err error) context.Context {
		if e := sqlutil.MapError(err); e != nil && e.Detail != "" {
			ctx = log.WithField(ctx, "SQLErrDetails", e.Detail)
		}

		return ctx
	})

	app := &App{
		l:      l,
		db:     db,
		pgx:    pool,
		cfg:    c,
		doneCh: make(chan struct{}),
		Logger: c.Logger,
	}

	if c.StatusAddr != "" {
		err = listenStatus(c.StatusAddr, app.doneCh)
		if err != nil {
			return nil, errors.Wrap(err, "start status listener")
		}
	}

	app.db.SetMaxIdleConns(c.DBMaxIdle)
	app.db.SetMaxOpenConns(c.DBMaxOpen)

	app.mgr = lifecycle.NewManager(app._Run, app._Shutdown)
	err = app.mgr.SetStartupFunc(app.startup)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// WaitForStartup will wait until the startup sequence is completed or the context is expired.
func (a *App) WaitForStartup(ctx context.Context) error {
	return a.mgr.WaitForStartup(a.Context(ctx))
}

// DB returns the sql.DB instance used by the application.
func (a *App) DB() *sql.DB { return a.db }

// URL returns the non-TLS listener URL of the application.
func (a *App) URL() string {
	return "http://" + a.l.Addr().String()
}

func (a *App) SMTPAddr() string {
	if a.smtpsrvL == nil {
		return ""
	}

	return a.smtpsrvL.Addr().String()
}
