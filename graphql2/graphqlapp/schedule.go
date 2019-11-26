package graphqlapp

import (
	context "context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type Schedule App

func (a *App) Schedule() graphql2.ScheduleResolver { return (*Schedule)(a) }

func (q *Query) Schedule(ctx context.Context, id string) (*schedule.Schedule, error) {
	return (*App)(q).FindOneSchedule(ctx, id)
}
func (s *Schedule) Shifts(ctx context.Context, raw *schedule.Schedule, start, end time.Time) ([]oncall.Shift, error) {
	if end.Before(start) {
		return nil, validation.NewFieldError("EndTime", "must be after StartTime")
	}
	if end.After(start.AddDate(0, 0, 50)) {
		return nil, validation.NewFieldError("EndTime", "cannot be more than 50 days past StartTime")
	}
	return s.OnCallStore.HistoryBySchedule(ctx, raw.ID, start, end)
}

func (s *Schedule) Target(ctx context.Context, raw *schedule.Schedule, input assignment.RawTarget) (*graphql2.ScheduleTarget, error) {
	rules, err := s.RuleStore.FindByTargetTx(ctx, nil, raw.ID, input)
	if err != nil {
		return nil, err
	}

	return &graphql2.ScheduleTarget{
		ScheduleID: raw.ID,
		Target:     &input,
		Rules:      rules,
	}, nil
}

func (s *Schedule) Targets(ctx context.Context, raw *schedule.Schedule) ([]graphql2.ScheduleTarget, error) {
	rules, err := s.RuleStore.FindAll(ctx, raw.ID)
	if err != nil {
		return nil, err
	}

	m := make(map[assignment.RawTarget][]rule.Rule)
	for _, r := range rules {
		tgt := assignment.RawTarget{ID: r.Target.TargetID(), Type: r.Target.TargetType()}
		m[tgt] = append(m[tgt], r)
	}

	result := make([]graphql2.ScheduleTarget, 0, len(m))
	for tgt, rules := range m {
		t := tgt // need to make a copy so we can take a pointer
		result = append(result, graphql2.ScheduleTarget{
			Target:     &t,
			ScheduleID: raw.ID,
			Rules:      rules,
		})
	}

	return result, nil
}
func (s *Schedule) AssignedTo(ctx context.Context, raw *schedule.Schedule) ([]assignment.RawTarget, error) {
	pols, err := s.PolicyStore.FindAllPoliciesBySchedule(ctx, raw.ID)
	if err != nil {
		return nil, err
	}
	sort.Slice(pols, func(i, j int) bool { return strings.ToLower(pols[i].Name) < strings.ToLower(pols[j].Name) })

	tgt := make([]assignment.RawTarget, len(pols))
	for i, p := range pols {
		tgt[i] = assignment.RawTarget{
			ID:   p.ID,
			Name: p.Name,
			Type: assignment.TargetTypeEscalationPolicy,
		}
	}

	return tgt, nil
}
func (m *Mutation) UpdateSchedule(ctx context.Context, input graphql2.UpdateScheduleInput) (ok bool, err error) {
	var loc *time.Location
	if input.TimeZone != nil {
		loc, err = util.LoadLocation(*input.TimeZone)
		if err != nil {
			return false, validation.NewFieldError("timeZone", err.Error())
		}
	}
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		sched, err := m.ScheduleStore.FindOneForUpdate(ctx, tx, input.ID)
		if err != nil {
			return err
		}
		if input.Name != nil {
			sched.Name = *input.Name
		}
		if input.Description != nil {
			sched.Description = *input.Description
		}

		if loc != nil {
			sched.TimeZone = loc
		}

		return m.ScheduleStore.UpdateTx(ctx, tx, sched)
	})

	return err == nil, err
}

func (m *Mutation) CreateSchedule(ctx context.Context, input graphql2.CreateScheduleInput) (sched *schedule.Schedule, err error) {
	usedTargets := make(map[assignment.RawTarget]int, len(input.Targets))

	for i, tgt := range input.Targets {
		fieldPrefix := fmt.Sprintf("targets[%d].", i)

		// validating both are not nil
		if tgt.NewRotation == nil && tgt.Target == nil {
			return nil, validate.Many(
				validation.NewFieldError(fieldPrefix+"target", "one of `target` or `newRotation` is required"),
				validation.NewFieldError(fieldPrefix+"newRotation", "one of `target` or `newRotation` is required"),
			)
		}

		// validating only one is present
		if tgt.NewRotation != nil && tgt.Target != nil {
			return nil, validate.Many(
				validation.NewFieldError(fieldPrefix+"target", "cannot be used with `newRotation`"),
				validation.NewFieldError(fieldPrefix+"newRotation", "cannot be used with `target`"),
			)
		}

		// checking for duplicate targets
		if tgt.Target != nil {
			raw := assignment.NewRawTarget(tgt.Target)
			if oldIndex, ok := usedTargets[raw]; ok {
				return nil, validation.NewFieldError(fieldPrefix+"target", fmt.Sprintf("must be unique. Conflicts with existing `targets[%d].target`.", oldIndex))
			}
			usedTargets[raw] = i
		}

	}

	loc, err := util.LoadLocation(input.TimeZone)
	if err != nil {
		return nil, validation.NewFieldError("timeZone", err.Error())
	}

	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		s := &schedule.Schedule{
			Name:     input.Name,
			TimeZone: loc,
		}
		if input.Description != nil {
			s.Description = *input.Description
		}
		sched, err = m.ScheduleStore.CreateScheduleTx(ctx, tx, s)
		if err != nil {
			return err
		}
		if input.Favorite != nil && *input.Favorite {
			err = m.FavoriteStore.SetTx(ctx, tx, permission.UserID(ctx), assignment.ScheduleTarget(sched.ID))
			if err != nil {
				return err
			}
		}
		for i := range input.Targets {
			if input.Targets[i].NewRotation == nil {
				continue
			}
			rot, err := m.CreateRotation(ctx, *input.Targets[i].NewRotation)
			if err != nil {
				return validation.AddPrefix("targets["+strconv.Itoa(i)+"].newRotation.", err)
			}
			// Inserting newly created rotation as 'target' with it's corresponding rules
			input.Targets[i].Target = &assignment.RawTarget{Type: assignment.TargetTypeRotation, ID: rot.ID, Name: rot.Name}

		}

		for i, r := range input.Targets {
			r.ScheduleID = &sched.ID
			_, err = m.UpdateScheduleTarget(ctx, r)
			if err != nil {
				return validation.AddPrefix("targets["+strconv.Itoa(i)+"].", err)
			}
		}

		return nil
	})

	return sched, err
}

func (r *Schedule) TimeZone(ctx context.Context, data *schedule.Schedule) (string, error) {
	return data.TimeZone.String(), nil
}

func (q *Query) Schedules(ctx context.Context, opts *graphql2.ScheduleSearchOptions) (conn *graphql2.ScheduleConnection, err error) {
	if opts == nil {
		opts = &graphql2.ScheduleSearchOptions{}
	}
	var searchOpts schedule.SearchOptions
	searchOpts.FavoritesUserID = permission.UserID(ctx)
	if opts.Search != nil {
		searchOpts.Search = *opts.Search
	}
	if opts.FavoritesOnly != nil {
		searchOpts.FavoritesOnly = *opts.FavoritesOnly
	}
	if opts.FavoritesFirst != nil {
		searchOpts.FavoritesFirst = *opts.FavoritesFirst
	}
	searchOpts.Omit = opts.Omit
	if opts.After != nil && *opts.After != "" {
		err = search.ParseCursor(*opts.After, &searchOpts)
		if err != nil {
			return nil, err
		}
	}
	if opts.First != nil {
		searchOpts.Limit = *opts.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	searchOpts.Limit++
	scheds, err := q.ScheduleStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.ScheduleConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(scheds) == searchOpts.Limit {
		scheds = scheds[:len(scheds)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(scheds) > 0 {
		last := scheds[len(scheds)-1]
		searchOpts.After.IsFavorite = last.IsUserFavorite()
		searchOpts.After.Name = last.Name

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = scheds
	return conn, err
}
func (s *Schedule) IsFavorite(ctx context.Context, raw *schedule.Schedule) (bool, error) {
	return raw.IsUserFavorite(), nil
}
