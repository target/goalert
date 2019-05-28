package graphql

import (
	"database/sql"
	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/engine/resolver"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/label"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/override"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/schedule/shiftcalc"
	"github.com/target/goalert/service"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/favorite"
	"github.com/target/goalert/user/notificationrule"
)

type Config struct {
	DB *sql.DB

	AlertStore    alert.Store
	AlertLogStore alertlog.Store
	UserStore     user.Store
	CMStore       contactmethod.Store
	NRStore       notificationrule.Store
	ServiceStore  service.Store

	ScheduleStore     schedule.Store
	ScheduleRuleStore rule.Store
	RotationStore     rotation.Store
	ShiftCalc         shiftcalc.Calculator

	EscalationStore     escalation.Store
	IntegrationKeyStore integrationkey.Store
	HeartbeatStore      heartbeat.Store

	LimitStore limit.Store

	OverrideStore override.Store

	Resolver          resolver.Resolver
	NotificationStore notification.Store
	UserFavoriteStore favorite.Store
	LabelStore        label.Store
	OnCallStore       oncall.Store
}
