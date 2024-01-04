package universalapi

import (
	"github.com/target/goalert/alert"
	"github.com/target/goalert/integrationkey"
)

// Config contains the values needed to implement the universal API handler.
type Config struct {
	AlertStore          *alert.Store
	IntegrationKeyStore *integrationkey.Store
	// ServiceRuleStore    *rule.Store
}
