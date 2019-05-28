package escalation

import (
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/validation/validate"
	"time"
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
	ID           string `json:"id"`
	PolicyID     string `json:"escalation_policy_id"`
	DelayMinutes int    `json:"delay_minutes"`
	StepNumber   int    `json:"step_number"`

	Targets []assignment.Target
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
