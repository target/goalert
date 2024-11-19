package engine

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/auth/authlink"
	"github.com/target/goalert/config"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
)

// Config contains parameters for controlling how the Engine operates.
type Config struct {
	AlertLogStore       *alertlog.Store
	AlertStore          *alert.Store
	ContactMethodStore  *contactmethod.Store
	NotificationManager *notification.Manager
	UserStore           *user.Store
	NotificationStore   *notification.Store
	NCStore             *notificationchannel.Store
	OnCallStore         *oncall.Store
	ScheduleStore       *schedule.Store
	AuthLinkStore       *authlink.Store
	SlackStore          *slack.ChannelSender
	DestRegistry        *nfydest.Registry
	River               *river.Client[pgx.Tx]
	RiverWorkers        *river.Workers

	ConfigSource config.Source

	Keys keyring.Keys

	MaxMessages int

	DisableCycle bool
	LogCycles    bool

	CycleTime time.Duration
}
