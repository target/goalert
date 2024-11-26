package graphqlapp

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/target/goalert/assignment"
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
		input.Actions[i], err = (*App)(a).CompatTargetToDest(ctx, tgt)
		if err != nil {
			return validation.NewFieldError(fmt.Sprintf("Targets[%d]", i), err.Error())
		}
	}

	return nil
}

func (a *App) UpdateEscalationPolicyStepInput() graphql2.UpdateEscalationPolicyStepInputResolver {
	return (*UpdateEscalationPolicyStepInput)(a)
}

func (a *UpdateEscalationPolicyStepInput) Targets(ctx context.Context, input *graphql2.UpdateEscalationPolicyStepInput, targets []assignment.RawTarget) error {
	input.Actions = make([]gadb.DestV1, len(targets))
	for i, tgt := range targets {
		var err error
		input.Actions[i], err = (*App)(a).CompatTargetToDest(ctx, tgt)
		if err != nil {
			return validation.NewFieldError(fmt.Sprintf("Targets[%d]", i), err.Error())
		}
	}

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
	if input.Actions != nil {
		// validate delay so we return a new coded error (when using actions)
		err := validate.Range("input.delayMinutes", input.DelayMinutes, 1, 9000)
		if err != nil {
			addInputError(ctx, err)
			return nil, errAlreadySet
		}
	}
	if len(input.Actions) != 0 && input.NewRotation != nil {
		return nil, validate.Many(
			validation.NewFieldError("actions", "cannot be used with `newRotation`"),
			validation.NewFieldError("newRotation", "cannot be used with `targets`"),
		)
	}

	if len(input.Actions) != 0 && input.NewSchedule != nil {
		return nil, validate.Many(
			validation.NewFieldError("actions", "cannot be used with `newSchedule`"),
			validation.NewFieldError("newSchedule", "cannot be used with `targets`"),
		)
	}

	if input.NewSchedule != nil && input.NewRotation != nil {
		return nil, validate.Many(
			validation.NewFieldError("newSchedule", "cannot be used with `newRotation`"),
			validation.NewFieldError("newRotation", "cannot be used with `newSchedule`"),
		)
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

		if input.NewRotation != nil {
			rot, err := m.CreateRotation(ctx, *input.NewRotation)
			if err != nil {
				return validation.AddPrefix("newRotation.", err)
			}

			// Should add to escalation_policy_actions
			err = m.PolicyStore.AddStepActionTx(ctx, tx, step.ID, rotation.DestFromID(rot.ID))
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
			err = m.PolicyStore.AddStepActionTx(ctx, tx, step.ID, schedule.DestFromID(sched.ID))
			if err != nil {
				return validation.AddPrefix("newSchedule.", err)
			}
		}

		userID := permission.UserID(ctx)
		for i, action := range input.Actions {
			if action.Type == user.DestTypeUser && action.Arg(user.FieldUserID) == "__current_user" {
				action.SetArg(user.FieldUserID, userID)
			}
			err = m.PolicyStore.AddStepActionTx(ctx, tx, step.ID, action)
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
			err = m.FavoriteStore.Set(ctx, gadb.Compat(tx), permission.UserID(ctx), assignment.EscalationPolicyTarget(pol.ID))
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

			inputStepIDs, err := validate.ParseManyUUID("stepIDs", input.StepIDs, len(steps))
			if err != nil {
				return err
			}

			// get list of step ids
			var stepIDs []uuid.UUID
			for _, step := range steps {
				stepIDs = append(stepIDs, step.ID)
			}

			// delete existing id if not found in input steps slice
			for _, stepID := range stepIDs {
				if !slices.Contains(inputStepIDs, stepID) {
					_, err = m.PolicyStore.DeleteStepTx(ctx, tx, stepID)
					if err != nil {
						return err
					}
				}
			}

			// loop through input steps to update order
			for i, stepID := range inputStepIDs {
				if !slices.Contains(stepIDs, stepID) {
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
		if input.Actions != nil {
			// get current actions
			existing, err := m.PolicyStore.FindAllStepActionsTx(ctx, gadb.Compat(tx), step.ID)
			if err != nil {
				return err
			}

			// We need to delete first, in case we're at the current system limit, that way the total number never exceeds the limit (unless the user is trying to add more than the limit).
			for _, action := range existing {
				stillWanted := slices.ContainsFunc(input.Actions, func(a gadb.DestV1) bool {
					return reflect.DeepEqual(a, action)
				})
				if stillWanted {
					// leave it alone
					continue
				}

				err = m.PolicyStore.DeleteStepActionTx(ctx, tx, step.ID, action)
				if err != nil {
					return err
				}
			}

			for _, action := range input.Actions {
				alreadyExists := slices.ContainsFunc(existing, func(e gadb.DestV1) bool {
					return reflect.DeepEqual(e, action)
				})
				if alreadyExists {
					// already exists, skip
					continue
				}

				err = m.PolicyStore.AddStepActionTx(ctx, tx, step.ID, action)
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
	return a.PolicyStore.FindAllStepActionsTx(ctx, nil, raw.ID)
}

func (step *EscalationPolicyStep) Targets(ctx context.Context, raw *escalation.Step) ([]assignment.RawTarget, error) {
	act, err := step.PolicyStore.FindAllStepActionsTx(ctx, nil, raw.ID)
	if err != nil {
		return nil, err
	}

	var targets []assignment.RawTarget
	for _, action := range act {
		tgt, err := CompatDestToTarget(action)
		if err != nil {
			return nil, err
		}

		targets = append(targets, tgt)
	}

	return targets, nil
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
