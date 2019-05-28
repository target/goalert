package limit

import (
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// Error represents an error caused by
type Error interface {
	error
	Max() int
	ID() ID
	Limit() bool
}

type limitErr struct {
	id  ID
	max int
}

var _ Error = &limitErr{}

// IsLimitError will determine if an error's cause is a limit.Error.
func IsLimitError(err error) bool {
	if e, ok := errors.Cause(err).(Error); ok && e.Limit() {
		return true
	}
	return false
}

// MapError will map a Postgres error that is caused by a limit constraint.
// If the given error is not caused by a known system limit constraint, nil is returned.
func MapError(err error) Error {
	e, ok := err.(*pq.Error)
	if !ok {
		return nil
	}
	if !strings.HasPrefix(e.Hint, "max=") {
		return nil
	}
	if !strings.HasSuffix(e.Constraint, "_limit") {
		return nil
	}
	id := ID(strings.TrimSuffix(e.Constraint, "_limit"))
	if id.Valid() != nil {
		return nil
	}
	m, err := strconv.Atoi(strings.TrimPrefix(e.Hint, "max="))
	if err != nil {
		return nil
	}
	return &limitErr{id: id, max: m}
}

func (l *limitErr) ClientError() bool { return true }

func (l *limitErr) Limit() bool { return true }
func (l *limitErr) ID() ID      { return l.id }
func (l *limitErr) Max() int    { return l.max }
func (l *limitErr) Error() string {
	switch l.id {
	case NotificationRulesPerUser:
		return "too many notification rules"
	case ContactMethodsPerUser:
		return "too many contact methods"
	case EPStepsPerPolicy:
		return "too many steps on this policy"
	case EPActionsPerStep:
		return "too many actions on this step"
	case ParticipantsPerRotation:
		return "too many participants on this rotation"
	case RulesPerSchedule:
		return "too many rules on this schedule"
	case IntegrationKeysPerService:
		return "too many integration keys on this service"
	case UnackedAlertsPerService:
		return "too many unacknowledged alerts for this service"
	case TargetsPerSchedule:
		return "too many targets on this schedule"
	case HeartbeatMonitorsPerService:
		return "too many heartbeat monitors on this service"
	case UserOverridesPerSchedule:
		return "too many user overrides on this schedule"
	}

	return "exceeded limit"
}
