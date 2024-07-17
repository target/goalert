package escalation

import (
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/validation/validate"
)

type ActiveStep struct {
	StepID          string
	PolicyID        string
	AlertID         int
	LastEscalation  time.Time
	LoopCount       int
	ForceEscalation bool
	StepNumber      int
}

type Step struct {
	ID           uuid.UUID `json:"id"`
	PolicyID     string    `json:"escalation_policy_id"`
	DelayMinutes int       `json:"delay_minutes"`
	StepNumber   int       `json:"step_number"`
}

func (s Step) Delay() time.Duration {
	return time.Duration(s.DelayMinutes) * time.Minute
}

func (s Step) Normalize() (*Step, error) {
	err := validate.Many(
		validate.UUID("PolicyID", s.PolicyID),
		validate.Range("DelayMinutes", s.DelayMinutes, 1, 9000),
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
