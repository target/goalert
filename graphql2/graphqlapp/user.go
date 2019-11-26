package graphqlapp

import (
	context "context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/notificationrule"
)

type User App

func (a *App) User() graphql2.UserResolver { return (*User)(a) }

func (a *User) AuthSubjects(ctx context.Context, obj *user.User) ([]user.AuthSubject, error) {
	return a.UserStore.FindAllAuthSubjectsForUser(ctx, obj.ID)
}
func (a *User) Role(ctx context.Context, usr *user.User) (graphql2.UserRole, error) {
	return graphql2.UserRole(usr.Role), nil
}

func (a *User) ContactMethods(ctx context.Context, obj *user.User) ([]contactmethod.ContactMethod, error) {
	return a.CMStore.FindAll(ctx, obj.ID)
}
func (a *User) NotificationRules(ctx context.Context, obj *user.User) ([]notificationrule.NotificationRule, error) {
	return a.NRStore.FindAll(ctx, obj.ID)
}

func (a *User) OnCallSteps(ctx context.Context, obj *user.User) ([]escalation.Step, error) {
	return a.PolicyStore.FindAllOnCallStepsForUserTx(ctx, nil, obj.ID)
}

func (a *Mutation) DeleteUser(ctx context.Context, id string) (bool, error) {
	err := a.UserStore.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	return true, nil
}
func (a *Mutation) UpdateUser(ctx context.Context, input graphql2.UpdateUserInput) (bool, error) {
	err := withContextTx(ctx, a.DB, func(ctx context.Context, tx *sql.Tx) error {
		usr, err := a.UserStore.FindOneTx(ctx, tx, input.ID, true)
		if err != nil {
			return err
		}
		if input.Name != nil {
			usr.Name = *input.Name
		}
		if input.Role != nil {
			usr.Role = permission.Role(*input.Role)
		}
		if input.Email != nil {
			usr.Email = *input.Email
		}
		if input.StatusUpdateContactMethodID != nil {
			usr.AlertStatusCMID = *input.StatusUpdateContactMethodID
		}
		return a.UserStore.UpdateTx(ctx, tx, usr)
	})
	return err == nil, err
}

func (q *Query) Users(ctx context.Context, opts *graphql2.UserSearchOptions, first *int, after, searchStr *string) (conn *graphql2.UserConnection, err error) {
	if opts == nil {
		opts = &graphql2.UserSearchOptions{
			First:  first,
			After:  after,
			Search: searchStr,
		}
	}

	var searchOpts user.SearchOptions
	if opts.Search != nil {
		searchOpts.Search = *opts.Search
	}
	searchOpts.Omit = opts.Omit
	if opts.After != nil && *opts.After != "" {
		err = search.ParseCursor(*opts.After, &searchOpts)
		if err != nil {
			return nil, errors.Wrap(err, "parse cursor")
		}
	}
	if opts.First != nil {
		searchOpts.Limit = *opts.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	searchOpts.Limit++
	users, err := q.UserStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.UserConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(users) == searchOpts.Limit {
		users = users[:len(users)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(users) > 0 {
		last := users[len(users)-1]
		searchOpts.After.Name = last.Name

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = users
	return conn, err
}

func (a *Query) User(ctx context.Context, id *string) (*user.User, error) {
	var userID string
	if id != nil {
		userID = *id
	} else {
		userID = permission.UserID(ctx)
	}
	return (*App)(a).FindOneUser(ctx, userID)
}
