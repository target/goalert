package signalapi

import (
	"database/sql"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/service/rule"
	"github.com/target/goalert/signal"
	"github.com/target/goalert/user"
)

// Config contains the values needed to implement the generic API handler.
type Config struct {
	DB *sql.DB

	AlertStore          *alert.Store
	SignalStore         *signal.Store
	ServiceRuleStore    *rule.Store
	IntegrationKeyStore *integrationkey.Store
	HeartbeatStore      *heartbeat.Store
	UserStore           *user.Store
}
