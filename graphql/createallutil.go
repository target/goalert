package graphql

import (
	"context"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/heartbeat"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/service"
	"github.com/target/goalert/validation"
)

type createAllData struct {
	EscalationPolicies    []escalation.Policy             `json:"escalation_policies"`
	EscalationPolicySteps []escalation.Step               `json:"escalation_policy_steps"`
	Services              []service.Service               `json:"services"`
	IntegrationKeys       []integrationkey.IntegrationKey `json:"integration_keys"`
	Rotations             []rotation.Rotation             `json:"rotations"`
	RotationParticipants  []rotation.Participant          `json:"rotation_participants"`
	Schedules             []schedule.Schedule             `json:"schedules"`
	ScheduleRules         []rule.Rule                     `json:"schedule_rules"`
	HeartbeatMonitors     []heartbeat.Monitor             `json:"heartbeat_monitors"`
	UserOverrides         []override.UserOverride         `json:"user_overrides"`
}

func (c *Config) createAll(ctx context.Context, data *createAllData) (*createAllData, error) {
	ids := make(map[string]string)
	setID := func(s, v string) error {
		if s == "" {
			return nil
		}
		if _, ok := ids[s]; ok {
			return validation.NewFieldError("duplicate value '%s'", s)
		}
		ids[s] = v
		return nil
	}
	setID("__current_user", permission.UserID(ctx))

	getID := func(s string) string {
		if s == "" {
			return ""
		}
		id, ok := ids[s]
		if ok {
			return id
		}
		return s
	}

	getTarget := func(tgt assignment.Target) assignment.Target {
		id, ok := ids[tgt.TargetID()]
		if ok {
			return &assignment.RawTarget{
				ID:   id,
				Type: tgt.TargetType(),
			}
		}
		return tgt
	}

	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var result createAllData

	// create escalation policies
	for _, ep := range data.EscalationPolicies {
		newEP, err := c.EscalationStore.CreatePolicyTx(ctx, tx, &ep)
		if err != nil {
			return nil, err
		}
		err = setID(ep.ID, newEP.ID)
		if err != nil {
			return nil, err
		}
		result.EscalationPolicies = append(result.EscalationPolicies, *newEP)
	}

	// create services
	for _, serv := range data.Services {
		serv.EscalationPolicyID = getID(serv.EscalationPolicyID)
		newServ, err := c.ServiceStore.CreateServiceTx(ctx, tx, &serv)
		if err != nil {
			return nil, err
		}
		err = setID(serv.ID, newServ.ID)
		if err != nil {
			return nil, err
		}
		result.Services = append(result.Services, *newServ)
	}

	// create integration keys
	for _, key := range data.IntegrationKeys {
		key.ServiceID = getID(key.ServiceID)
		newKey, err := c.IntegrationKeyStore.CreateKeyTx(ctx, tx, &key)
		if err != nil {
			return nil, err
		}
		result.IntegrationKeys = append(result.IntegrationKeys, *newKey)
	}

	// create heartbeat monitors
	for _, hb := range data.HeartbeatMonitors {
		hb.ServiceID = getID(hb.ServiceID)
		newHB, err := c.HeartbeatStore.CreateTx(ctx, tx, &hb)
		if err != nil {
			return nil, err
		}
		result.HeartbeatMonitors = append(result.HeartbeatMonitors, *newHB)
	}

	// create rotations
	for _, rot := range data.Rotations {
		newRot, err := c.RotationStore.CreateRotationTx(ctx, tx, &rot)
		if err != nil {
			return nil, err
		}
		err = setID(rot.ID, newRot.ID)
		if err != nil {
			return nil, err
		}
		result.Rotations = append(result.Rotations, *newRot)
	}

	// add rotation participants
	for _, rp := range data.RotationParticipants {
		rp.RotationID = getID(rp.RotationID)
		rp.Target = getTarget(rp.Target)

		newRP, err := c.RotationStore.AddParticipantTx(ctx, tx, &rp)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		result.RotationParticipants = append(result.RotationParticipants, *newRP)
	}

	// create schedules
	for _, sched := range data.Schedules {
		newSched, err := c.ScheduleStore.CreateScheduleTx(ctx, tx, &sched)
		if err != nil {
			return nil, err
		}
		err = setID(sched.ID, newSched.ID)
		if err != nil {
			return nil, err
		}
		result.Schedules = append(result.Schedules, *newSched)
	}

	// create user overrides
	for _, o := range data.UserOverrides {
		o.AddUserID = getID(o.AddUserID)
		o.RemoveUserID = getID(o.RemoveUserID)
		o.Target = getTarget(o.Target)
		newO, err := c.OverrideStore.CreateUserOverrideTx(ctx, tx, &o)
		if err != nil {
			return nil, err
		}
		result.UserOverrides = append(result.UserOverrides, *newO)
	}

	// create rules for schedule(s, depending on placeholder ids)
	for _, r := range data.ScheduleRules {
		r.ScheduleID = getID(r.ScheduleID)
		r.Target = getTarget(r.Target)

		newRule, err := c.ScheduleRuleStore.CreateRuleTx(ctx, tx, &r)
		if err != nil {
			return nil, err
		}
		result.ScheduleRules = append(result.ScheduleRules, *newRule)
	}

	// create steps for escalation policy(s)
	for _, step := range data.EscalationPolicySteps {
		step.PolicyID = getID(step.PolicyID)
		newStep, err := c.EscalationStore.CreateStepTx(ctx, tx, &step)
		if err != nil {
			return nil, err
		}
		result.EscalationPolicySteps = append(result.EscalationPolicySteps, *newStep)

		for _, tgt := range step.Targets {
			tgt = getTarget(tgt)
			err = c.EscalationStore.AddStepTargetTx(ctx, tx, newStep.ID, tgt)
			if err != nil {
				return nil, err
			}
		}
	}

	return &result, tx.Commit()
}
