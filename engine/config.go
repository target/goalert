package engine

import (
	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/config"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
)

// Config contains parameters for controlling how the Engine operates.
type Config struct {
	AlertlogStore      alertlog.Store
	AlertStore         alert.Store
	ContactMethodStore contactmethod.Store
	NotificationSender notification.Sender
	UserStore          user.Store
	NotificationStore  notification.Store
	NCStore            notificationchannel.Store

	ConfigSource config.Source

	Keys keyring.Keys

	MaxMessages int

	DisableCycle bool
}
