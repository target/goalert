package graphqlapp

import (
	context "context"
	"database/sql"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/override"
	"github.com/target/goalert/search"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"
)

type UserOverride App

func (a *App) UserOverride() graphql2.UserOverrideResolver { return (*UserOverride)(a) }
func (q *Query) UserOverride(ctx context.Context, id string) (*override.UserOverride, error) {
	return q.OverrideStore.FindOneUserOverrideTx(ctx, nil, id, false)
}

func (m *Mutation) UpdateUserOverride(ctx context.Context, input graphql2.UpdateUserOverrideInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		u, err := m.OverrideStore.FindOneUserOverrideTx(ctx, tx, input.ID, true)
		if err != nil {
			return err
		}
		if u == nil {
			return validation.NewFieldError("ID", "user override not found")
		}

		if input.Start != nil {
			u.Start = *input.Start
		}
		if input.End != nil {
			u.End = *input.End
		}
		if input.AddUserID != nil {
			u.AddUserID = *input.AddUserID
		}
		if input.RemoveUserID != nil {
			u.RemoveUserID = *input.RemoveUserID
		}

		return m.OverrideStore.UpdateUserOverrideTx(ctx, tx, u)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *Mutation) CreateUserOverride(ctx context.Context, input graphql2.CreateUserOverrideInput) (*override.UserOverride, error) {
	if input.ScheduleID == nil {
		return nil, validation.NewFieldError("ScheduleID", "is required")
	}
	u := &override.UserOverride{
		Target: assignment.ScheduleTarget(*input.ScheduleID),
		Start:  input.Start,
		End:    input.End,
	}
	if input.AddUserID != nil {
		u.AddUserID = *input.AddUserID
	}
	if input.RemoveUserID != nil {
		u.RemoveUserID = *input.RemoveUserID
	}
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		u, err = m.OverrideStore.CreateUserOverrideTx(ctx, tx, u)
		return err
	})
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (u *UserOverride) AddUser(ctx context.Context, raw *override.UserOverride) (*user.User, error) {
	if raw.AddUserID == "" {
		return nil, nil
	}
	return (*App)(u).FindOneUser(ctx, raw.AddUserID)
}

func (u *UserOverride) RemoveUser(ctx context.Context, raw *override.UserOverride) (*user.User, error) {
	if raw.RemoveUserID == "" {
		return nil, nil
	}
	return (*App)(u).FindOneUser(ctx, raw.RemoveUserID)
}

func (u *UserOverride) Target(ctx context.Context, raw *override.UserOverride) (*assignment.RawTarget, error) {
	tgt := assignment.NewRawTarget(raw.Target)
	return &tgt, nil
}

func (q *Query) UserOverrides(ctx context.Context, input *graphql2.UserOverrideSearchOptions) (conn *graphql2.UserOverrideConnection, err error) {
	if input == nil {
		input = &graphql2.UserOverrideSearchOptions{}
	}

	var searchOpts override.SearchOptions
	searchOpts.Omit = input.Omit
	if input.After != nil && *input.After != "" {
		err = search.ParseCursor(*input.After, &searchOpts)
		if err != nil {
			return nil, err
		}
	} else {
		searchOpts.AddUserIDs = input.FilterAddUserID
		searchOpts.RemoveUserIDs = input.FilterRemoveUserID
		searchOpts.AnyUserIDs = input.FilterAnyUserID
		if input.ScheduleID != nil {
			searchOpts.ScheduleID = *input.ScheduleID
		}
		if input.Start != nil {
			searchOpts.Start = *input.Start
		}
		if input.End != nil {
			searchOpts.End = *input.End
		}
	}
	if input.First != nil {
		searchOpts.Limit = *input.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	searchOpts.Limit++
	overrides, err := q.OverrideStore.Search(ctx, q.DBTX, &searchOpts)
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.UserOverrideConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(overrides) == searchOpts.Limit {
		overrides = overrides[:len(overrides)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(overrides) > 0 {
		last := overrides[len(overrides)-1]
		searchOpts.After.ID = last.ID

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = overrides
	return conn, err
}
