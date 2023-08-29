package notifyapi

import (
	"github.com/target/goalert/alert"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/user"
)

// Config contains the values needed to implement the generic API handler.
type Config struct {
	AlertStore          *alert.Store
	IntegrationKeyStore *integrationkey.Store
	HeartbeatStore      *heartbeat.Store
	UserStore           *user.Store
}
