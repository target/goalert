package errutil

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/util/sqlutil"
)

// LimitError represents an error caused by a configured system limit.
type LimitError interface {
	error
	Max() int
	ID() limit.ID
	Limit() bool
}

type limitErr struct {
	id  limit.ID
	max int
}

var _ LimitError = &limitErr{}

// IsLimitError will determine if an error's cause is a limit.Error.
func IsLimitError(err error) bool {
	var e LimitError
	if errors.As(err, &e) && e.Limit() {
		return true
	}
	return false
}

func mapLimitError(s *sqlutil.Error) LimitError {
	if s == nil {
		return nil
	}
	if !strings.HasPrefix(s.Hint, "max=") {
		return nil
	}
	if !strings.HasSuffix(s.ConstraintName, "_limit") {
		return nil
	}
	id := limit.ID(strings.TrimSuffix(s.ConstraintName, "_limit"))
	if id.Valid() != nil {
		return nil
	}
	m, err := strconv.Atoi(strings.TrimPrefix(s.Hint, "max="))
	if err != nil {
		return nil
	}
	return &limitErr{id: id, max: m}
}

func (l *limitErr) ClientError() bool { return true }

func (l *limitErr) Limit() bool  { return true }
func (l *limitErr) ID() limit.ID { return l.id }
func (l *limitErr) Max() int     { return l.max }
func (l *limitErr) Error() string {
	switch l.id {
	case limit.NotificationRulesPerUser:
		return "too many notification rules"
	case limit.ContactMethodsPerUser:
		return "too many contact methods"
	case limit.EPStepsPerPolicy:
		return "too many steps on this policy"
	case limit.EPActionsPerStep:
		return "too many actions on this step"
	case limit.ParticipantsPerRotation:
		return "too many participants on this rotation"
	case limit.RulesPerSchedule:
		return "too many rules on this schedule"
	case limit.IntegrationKeysPerService:
		return "too many integration keys on this service"
	case limit.UnackedAlertsPerService:
		return "too many unacknowledged alerts for this service"
	case limit.TargetsPerSchedule:
		return "too many targets on this schedule"
	case limit.HeartbeatMonitorsPerService:
		return "too many heartbeat monitors on this service"
	case limit.UserOverridesPerSchedule:
		return "too many user overrides on this schedule"
	case limit.CalendarSubscriptionsPerUser:
		return "too many calendar subscriptions for this user"
	case limit.PendingSignalsPerService:
		return "too many pending signals for this service"
	case limit.PendingSignalsPerDestPerService:
		return "too many pending signals for this destination on this service"
	}

	return "exceeded limit"
}
