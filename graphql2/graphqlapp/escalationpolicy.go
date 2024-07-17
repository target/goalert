package graphqlapp

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/config"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notice"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/search"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type (
	EscalationPolicy                App
	EscalationPolicyStep            App
	CreateEscalationPolicyStepInput App
	UpdateEscalationPolicyStepInput App
)

func (a *App) EscalationPolicy() graphql2.EscalationPolicyResolver { return (*EscalationPolicy)(a) }
func (a *App) EscalationPolicyStep() graphql2.EscalationPolicyStepResolver {
	return (*EscalationPolicyStep)(a)
}

func (a *App) CreateEscalationPolicyStepInput() graphql2.CreateEscalationPolicyStepInputResolver {
	return (*CreateEscalationPolicyStepInput)(a)
}

func (a *CreateEscalationPolicyStepInput) Targets(ctx context.Context, input *graphql2.CreateEscalationPolicyStepInput, targets []assignment.RawTarget) error {
	input.Actions = make([]gadb.DestV1, len(targets))
	for i, tgt := range targets {
		var err error
		input.Actions[i], err = CompatTargetToDest(tgt)
		if err != nil {
			return validation.NewFieldError(fmt.Sprintf("Targets[%d]", i), err.Error())
		}
	}

	return nil
}

func (a *App) UpdateEscalationPolicyStepInput() graphql2.UpdateEscalationPolicyStepInputResolver {
	return (*UpdateEscalationPolicyStepInput)(a)
}

func (a *UpdateEscalationPolicyStepInput) Actions(ctx context.Context, input *graphql2.UpdateEscalationPolicyStepInput, actions []gadb.DestV1) error {
	tgts := make([]assignment.RawTarget, len(actions))
	var err error
	for i, action := range actions {
		if err := (*App)(a).ValidateDestination(ctx, fmt.Sprintf("%d.dest", i), &action); err != nil {
			return err
		}
		tgts[i], err = CompatDestToTarget(action)
		if err != nil {
			return validation.NewFieldError("actions", "invalid DestInput")
		}
	}
	input.Targets = tgts
	return nil
}

func contains(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func (m *Mutation) CreateEscalationPolicyStep(ctx context.Context, input graphql2.CreateEscalationPolicyStepInput) (step *escalation.Step, err error) {
	cfg := config.FromContext(ctx)
	if input.Actions != nil {
		// validate delay so we return a new coded error (when using actions)
		err := validate.Range("input.delayMinutes", input.DelayMinutes, 1, 9000)
		if err != nil {
			addInputError(ctx, err)
			return nil, errAlreadySet
		}
	}
	if len(input.Targets) != 0 && input.NewRotation != nil {
		return nil, validate.Many(
			validation.NewFieldError("targets", "cannot be used with `newRotation`"),
			validation.NewFieldError("newRotation", "cannot be used with `targets`"),
		)
	}

	if len(input.Targets) != 0 && input.NewSchedule != nil {
		return nil, validate.Many(
			validation.NewFieldError("targets", "cannot be used with `newSchedule`"),
			validation.NewFieldError("newSchedule", "cannot be used with `targets`"),
		)
	}

	if input.NewSchedule != nil && input.NewRotation != nil {
		return nil, validate.Many(
			validation.NewFieldError("newSchedule", "cannot be used with `newRotation`"),
			validation.NewFieldError("newRotation", "cannot be used with `newSchedule`"),
		)
	}

	for _, tgt := range input.Targets {
		if tgt.Type == assignment.TargetTypeChanWebhook && !cfg.ValidWebhookURL(tgt.ID) {
			// UI code expects targets to be un-indexed
			return nil, validation.NewFieldError("targets", "URL not allowed by administrator")
		}
	}

	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		s := &escalation.Step{
			DelayMinutes: input.DelayMinutes,
		}
		if input.EscalationPolicyID != nil {
			s.PolicyID = *input.EscalationPolicyID
		}

		step, err = m.PolicyStore.CreateStepTx(ctx, tx, s)
		if err != nil {
			return err
		}
		stepID := uuid.MustParse(step.ID)

		if input.NewRotation != nil {
			rot, err := m.CreateRotation(ctx, *input.NewRotation)
			if err != nil {
				return validation.AddPrefix("newRotation.", err)
			}

			// Should add to escalation_policy_actions
			err = m.PolicyStore.AddStepActionTx(ctx, tx, stepID, rotation.DestFromID(rot.ID))
			if err != nil {
				return validation.AddPrefix("newRotation.", err)
			}
		}

		if input.NewSchedule != nil {
			sched, err := m.CreateSchedule(ctx, *input.NewSchedule)
			if err != nil {
				return validation.AddPrefix("newSchedule.", err)
			}

			// Should add to escalation_policy_actions
			err = m.PolicyStore.AddStepActionTx(ctx, tx, stepID, schedule.DestFromID(sched.ID))
			if err != nil {
				return validation.AddPrefix("newSchedule.", err)
			}
		}

		userID := permission.UserID(ctx)
		for i, action := range input.Actions {
			if action.Type == user.DestTypeUser && action.Arg(user.FieldUserID) == "__current_user" {
				action.SetArg(user.FieldUserID, userID)
			}
			err = m.PolicyStore.AddStepActionTx(ctx, tx, stepID, action)
			if err != nil {
				return validation.AddPrefix("Actions["+strconv.Itoa(i)+"].", err)
			}
		}

		return err
	})

	return step, err
}

func (m *Mutation) CreateEscalationPolicy(ctx context.Context, input graphql2.CreateEscalationPolicyInput) (pol *escalation.Policy, err error) {
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		p := &escalation.Policy{
			Name: input.Name,
		}
		if input.Repeat != nil {
			p.Repeat = *input.Repeat
		}
		if input.Description != nil {
			p.Description = *input.Description
		}

		pol, err = m.PolicyStore.CreatePolicyTx(ctx, tx, p)
		if err != nil {
			return err
		}
		if input.Favorite != nil && *input.Favorite {
			err = m.FavoriteStore.Set(ctx, tx, permission.UserID(ctx), assignment.EscalationPolicyTarget(pol.ID))
			if err != nil {
				return err
			}
		}

		for i, step := range input.Steps {
			step.EscalationPolicyID = &pol.ID
			_, err = m.CreateEscalationPolicyStep(ctx, step)
			if err != nil {
				return validation.AddPrefix("Steps["+strconv.Itoa(i)+"].", err)
			}
		}
		return err
	})

	return pol, err
}

func (m *Mutation) UpdateEscalationPolicy(ctx context.Context, input graphql2.UpdateEscalationPolicyInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		ep, err := m.PolicyStore.FindOnePolicyForUpdateTx(ctx, tx, input.ID)
		if err != nil {
			return err
		}

		if input.Name != nil {
			ep.Name = *input.Name
		}

		if input.Description != nil {
			ep.Description = *input.Description
		}

		if input.Repeat != nil {
			ep.Repeat = *input.Repeat
		}

		err = m.PolicyStore.UpdatePolicyTx(ctx, tx, ep)
		if err != nil {
			return err
		}

		if input.StepIDs != nil {
			// get current steps on policy
			steps, err := m.PolicyStore.FindAllStepsTx(ctx, tx, input.ID)
			if err != nil {
				return err
			}

			// get list of step ids
			var stepIDs []string
			for _, step := range steps {
				stepIDs = append(stepIDs, step.ID)
			}

			// delete existing id if not found in input steps slice
			for _, stepID := range stepIDs {
				if !contains(input.StepIDs, stepID) {
					_, err = m.PolicyStore.DeleteStepTx(ctx, tx, stepID)
					if err != nil {
						return err
					}
				}
			}

			// loop through input steps to update order
			for i, stepID := range input.StepIDs {
				if !contains(stepIDs, stepID) {
					return validation.NewFieldError("steps["+strconv.Itoa(i)+"]", "uuid does not exist on policy")
				}

				err = m.PolicyStore.UpdateStepNumberTx(ctx, tx, stepID, i)
				if err != nil {
					return err
				}
			}
		}

		return err
	})

	return true, err
}

func (m *Mutation) UpdateEscalationPolicyStep(ctx context.Context, input graphql2.UpdateEscalationPolicyStepInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		cfg := config.FromContext(ctx)
		step, err := m.PolicyStore.FindOneStepForUpdateTx(ctx, tx, input.ID) // get delay
		if err != nil {
			return err
		}

		// update delay if provided
		if input.DelayMinutes != nil {
			step.DelayMinutes = *input.DelayMinutes

			err = m.PolicyStore.UpdateStepDelayTx(ctx, tx, step.ID, step.DelayMinutes)
			if err != nil {
				return err
			}
		}

		// update targets if provided
		if input.Targets != nil {
			step.Targets = make([]assignment.Target, len(input.Targets))
			for i, tgt := range input.Targets {
				if tgt.Type == assignment.TargetTypeChanWebhook && !cfg.ValidWebhookURL(tgt.ID) {
					// UI code expects targets to be un-indexed
					return validation.NewFieldError("targets", "URL not allowed by administrator")
				}
				step.Targets[i] = tgt
			}

			// get current targets on step
			curr, err := m.PolicyStore.FindAllStepTargetsTx(ctx, tx, step.ID)
			if err != nil {
				return err
			}

			wantedTargets := make(map[assignment.RawTarget]int, len(step.Targets))
			currentTargets := make(map[assignment.RawTarget]bool, len(curr))

			// construct maps
			for i, tgt := range step.Targets {
				rt := assignment.NewRawTarget(tgt)
				if oldIdx, ok := wantedTargets[rt]; ok {
					return validation.NewFieldError(fmt.Sprintf("Targets[%d]", i), fmt.Sprintf("Duplicates existing target at index %d.", oldIdx))
				}
				wantedTargets[rt] = i
			}
			for _, tgt := range curr {
				currentTargets[assignment.NewRawTarget(tgt)] = true
			}

			// add targets in wanted that are not in curr
			for tgt, idx := range wantedTargets {
				if currentTargets[tgt] {
					continue
				}

				// add new step
				err = m.PolicyStore.AddStepTargetTx(ctx, tx, step.ID, tgt)
				if err != nil {
					return validation.AddPrefix(fmt.Sprintf("Targets[%d].", idx), err)
				}
			}

			// remove targets in curr that are not in wanted
			for tgt := range currentTargets {
				if _, ok := wantedTargets[tgt]; ok {
					continue
				}

				// delete unwanted step
				err = m.PolicyStore.DeleteStepTargetTx(ctx, tx, step.ID, tgt)
				if err != nil {
					return err
				}
			}
		}

		return err
	})

	return true, err
}

func (a *EscalationPolicyStep) Actions(ctx context.Context, raw *escalation.Step) ([]gadb.DestV1, error) {
	tgts, err := a.Targets(ctx, raw)
	if err != nil {
		return nil, err
	}

	actions := make([]gadb.DestV1, len(tgts))
	for i, tgt := range tgts {
		actions[i], err = CompatTargetToDest(tgt)
		if err != nil {
			return nil, err
		}
	}

	return actions, nil
}

func (step *EscalationPolicyStep) Targets(ctx context.Context, raw *escalation.Step) ([]assignment.RawTarget, error) {
	// TODO: use dataloader
	var targets []assignment.Target
	var err error
	if len(raw.Targets) > 0 {
		targets = raw.Targets
	} else {
		targets, err = step.PolicyStore.FindAllStepTargetsTx(ctx, nil, raw.ID)
		if err != nil {
			return nil, err
		}
	}

	result := make([]assignment.RawTarget, len(targets))
	for i, tgt := range targets {
		switch t := tgt.(type) {
		case *assignment.RawTarget:
			result[i] = *t
		case assignment.RawTarget:
			result[i] = t
		default:
			result[i] = assignment.NewRawTarget(t)
		}
	}

	return result, nil
}

func (step *EscalationPolicyStep) EscalationPolicy(ctx context.Context, raw *escalation.Step) (*escalation.Policy, error) {
	return (*App)(step).FindOnePolicy(ctx, raw.PolicyID)
}

func (step *EscalationPolicy) IsFavorite(ctx context.Context, raw *escalation.Policy) (bool, error) {
	return raw.IsUserFavorite(), nil
}

func (ep *EscalationPolicy) Steps(ctx context.Context, raw *escalation.Policy) ([]escalation.Step, error) {
	return ep.PolicyStore.FindAllSteps(ctx, raw.ID)
}

func (ep *EscalationPolicy) Notices(ctx context.Context, raw *escalation.Policy) ([]notice.Notice, error) {
	return ep.NoticeStore.FindAllPolicyNotices(ctx, raw.ID)
}

func (ep *EscalationPolicy) AssignedTo(ctx context.Context, raw *escalation.Policy) ([]assignment.RawTarget, error) {
	svcs, err := ep.ServiceStore.FindAllByEP(ctx, raw.ID)
	if err != nil {
		return nil, err
	}

	var tgts []assignment.RawTarget
	for _, svc := range svcs {
		var tgt assignment.RawTarget
		tgt.ID = svc.ID
		tgt.Name = svc.Name
		tgt.Type = assignment.TargetTypeService
		tgts = append(tgts, tgt)
	}

	return tgts, nil
}

func (q *Query) EscalationPolicy(ctx context.Context, id string) (*escalation.Policy, error) {
	return (*App)(q).FindOnePolicy(ctx, id)
}

func (q *Query) EscalationPolicies(ctx context.Context, opts *graphql2.EscalationPolicySearchOptions) (conn *graphql2.EscalationPolicyConnection, err error) {
	if opts == nil {
		opts = &graphql2.EscalationPolicySearchOptions{}
	}

	var searchOpts escalation.SearchOptions
	searchOpts.FavoritesUserID = permission.UserID(ctx)
	if opts.Search != nil {
		searchOpts.Search = *opts.Search
	}
	searchOpts.Omit = opts.Omit
	if opts.After != nil && *opts.After != "" {
		err = search.ParseCursor(*opts.After, &searchOpts)
		if err != nil {
			return nil, err
		}
	}
	if opts.FavoritesOnly != nil {
		searchOpts.FavoritesOnly = *opts.FavoritesOnly
	}
	if opts.FavoritesFirst != nil {
		searchOpts.FavoritesFirst = *opts.FavoritesFirst
	}
	if opts.First != nil {
		searchOpts.Limit = *opts.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	searchOpts.Limit++
	pols, err := q.PolicyStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.EscalationPolicyConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(pols) == searchOpts.Limit {
		pols = pols[:len(pols)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(pols) > 0 {
		last := pols[len(pols)-1]
		searchOpts.After.Name = last.Name

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return nil, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = pols
	return conn, err
}
