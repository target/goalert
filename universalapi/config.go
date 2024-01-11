package universalapi

import (
	"database/sql"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/integrationkey/integrationkeyrule"
)

// Config contains the values needed to implement the universal API handler.
type Config struct {
	AlertStore          *alert.Store
	IntegrationKeyStore *integrationkey.Store
	IntKeyRuleStore     *integrationkeyrule.Store
	DB                  *sql.DB
}
