package graphql

import (
	"context"
	"database/sql"
)

type deleteAllData struct {
	EscalationPolicyIDs     []string `json:"escalation_policie_ids"`
	EscalationPolicyStepIDs []string `json:"escalation_policy_step_ids"`
	ServiceIDs              []string `json:"service_ids"`
	IntegrationKeyIDs       []string `json:"integration_key_ids"`
	RotationIDs             []string `json:"rotation_ids"`
	RotationParticipantIDs  []string `json:"rotation_participant_ids"`
	ScheduleIDs             []string `json:"schedule_ids"`
	ScheduleRuleIDs         []string `json:"schedule_rule_ids"`
	HeartbeatMonitorIDs     []string `json:"heartbeat_monitor_ids"`
	UserOverrideIDs         []string `json:"user_override_ids"`
}

func (c *Config) deleteAll(ctx context.Context, data *deleteAllData) error {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteIDs := func(fn func(context.Context, *sql.Tx, string) error, ids []string) {
		if err != nil {
			return
		}
		for _, id := range ids {
			err = fn(ctx, tx, id)
			if err != nil {
				return
			}
		}
	}

	deleteIDs(func(ctx context.Context, tx *sql.Tx, id string) error {
		return c.OverrideStore.DeleteUserOverrideTx(ctx, tx, id)
	}, data.UserOverrideIDs)
	deleteIDs(c.HeartbeatStore.DeleteTx, data.HeartbeatMonitorIDs)
	deleteIDs(c.ScheduleRuleStore.DeleteTx, data.ScheduleRuleIDs)
	deleteIDs(c.IntegrationKeyStore.DeleteTx, data.IntegrationKeyIDs)
	deleteIDs(func(ctx context.Context, tx *sql.Tx, id string) error {
		_, err := c.EscalationStore.DeleteStepTx(ctx, tx, id)
		return err
	}, data.EscalationPolicyStepIDs)
	deleteIDs(func(ctx context.Context, tx *sql.Tx, id string) error {
		_, err := c.RotationStore.RemoveParticipantTx(ctx, tx, id)
		return err
	}, data.RotationParticipantIDs)
	deleteIDs(c.ScheduleStore.DeleteTx, data.ScheduleIDs)
	deleteIDs(c.RotationStore.DeleteRotationTx, data.RotationIDs)
	deleteIDs(c.ServiceStore.DeleteTx, data.ServiceIDs)
	deleteIDs(c.EscalationStore.DeletePolicyTx, data.EscalationPolicyIDs)

	if err != nil {
		return err
	}

	return tx.Commit()
}
