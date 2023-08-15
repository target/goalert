package graphqlapp

import (
	context "context"
	"database/sql"

	"github.com/target/goalert/auth"
	"github.com/target/goalert/auth/basic"
	"github.com/target/goalert/calsub"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/escalation"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/user/notificationrule"
)

type (
	User        App
	UserSession App
)

func (a *App) User() graphql2.UserResolver { return (*User)(a) }

func (a *App) UserSession() graphql2.UserSessionResolver { return (*UserSession)(a) }

func (a *User) Sessions(ctx context.Context, obj *user.User) ([]auth.UserSession, error) {
	return a.AuthHandler.FindAllUserSessions(ctx, obj.ID)
}

func (a *UserSession) Current(ctx context.Context, obj *auth.UserSession) (bool, error) {
	src := permission.Source(ctx)
	if src == nil {
		return false, nil
	}
	if src.Type != permission.SourceTypeAuthProvider {
		return false, nil
	}

	return obj.ID == src.ID, nil
}

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

func (a *User) CalendarSubscriptions(ctx context.Context, obj *user.User) ([]calsub.Subscription, error) {
	return a.CalSubStore.FindAllByUser(ctx, obj.ID)
}

func (a *User) OnCallSteps(ctx context.Context, obj *user.User) ([]escalation.Step, error) {
	return a.PolicyStore.FindAllOnCallStepsForUserTx(ctx, nil, obj.ID)
}

func (a *Mutation) CreateBasicAuth(ctx context.Context, input graphql2.CreateBasicAuthInput) (bool, error) {
	pw, err := a.AuthBasicStore.NewHashedPassword(ctx, input.Password)
	if err != nil {
		return false, err
	}

	err = withContextTx(ctx, a.DB, func(ctx context.Context, tx *sql.Tx) error {
		return a.AuthBasicStore.CreateTx(ctx, tx, input.UserID, input.Username, pw)
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (a *Mutation) UpdateBasicAuth(ctx context.Context, input graphql2.UpdateBasicAuthInput) (bool, error) {
	var validatedPW basic.ValidatedPassword
	var err error
	if input.OldPassword != nil {
		if *input.OldPassword == input.Password {
			return false, validation.NewFieldError("Password", "Cannot match OldPassword")
		}
		validatedPW, err = a.AuthBasicStore.ValidatePassword(ctx, *input.OldPassword)
		if err != nil {
			return false, err
		}
	}
	pw, err := a.AuthBasicStore.NewHashedPassword(ctx, input.Password)
	if err != nil {
		return false, err
	}

	err = withContextTx(ctx, a.DB, func(ctx context.Context, tx *sql.Tx) error {
		return a.AuthBasicStore.UpdateTx(ctx, tx, input.UserID, validatedPW, pw)
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (a *Mutation) CreateUser(ctx context.Context, input graphql2.CreateUserInput) (*user.User, error) {
	var newUser *user.User

	// NOTE input.username must be validated before input.name
	// user's name defaults to input.username and a user must be created before an auth_basic_user
	err := validate.Username("Username", input.Username)
	if err != nil {
		return nil, err
	}

	pass, err := a.AuthBasicStore.NewHashedPassword(ctx, input.Password)
	if err != nil {
		return nil, err
	}

	// user default values
	usr := &user.User{
		Name: input.Username,
		Role: permission.RoleUser,
	}

	if input.Name != nil {
		usr.Name = *input.Name
	}

	if input.Email != nil {
		usr.Email = *input.Email
	}

	if input.Role != nil {
		usr.Role = permission.Role(*input.Role)
	}

	err = withContextTx(ctx, a.DB, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		newUser, err = a.UserStore.InsertTx(ctx, tx, usr)
		if err != nil {
			return err
		}
		if input.Favorite != nil && *input.Favorite {
			err = a.FavoriteStore.SetTx(ctx, tx, permission.UserID(ctx), assignment.UserTarget(newUser.ID))
			if err != nil {
				return err
			}
		}
		err = a.AuthBasicStore.CreateTx(ctx, tx, newUser.ID, input.Username, pass)
		if err != nil {
			return err
		}
		return nil
	})

	return newUser, err
}

func (a *Mutation) UpdateUser(ctx context.Context, input graphql2.UpdateUserInput) (bool, error) {
	err := withContextTx(ctx, a.DB, func(ctx context.Context, tx *sql.Tx) error {
		usr, err := a.UserStore.FindOneTx(ctx, tx, input.ID, true)
		if err != nil {
			return err
		}

		if input.Role != nil {
			err = a.UserStore.SetUserRoleTx(ctx, tx, input.ID, permission.Role(*input.Role))
			if err != nil {
				return err
			}
		}

		if input.Name != nil {
			usr.Name = *input.Name
		}
		if input.Email != nil {
			usr.Email = *input.Email
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
	searchOpts.FavoritesUserID = permission.UserID(ctx)
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
	if opts.CMValue != nil {
		searchOpts.CMValue = *opts.CMValue
	}
	if opts.CMType != nil {
		searchOpts.CMType = *opts.CMType
	}
	if opts.FavoritesOnly != nil {
		searchOpts.FavoritesOnly = *opts.FavoritesOnly
	}
	if opts.FavoritesFirst != nil {
		searchOpts.FavoritesFirst = *opts.FavoritesFirst
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

func (a *User) IsFavorite(ctx context.Context, raw *user.User) (bool, error) {
	return raw.IsUserFavorite(), nil
}

func (a *Query) UserShifts(ctx context.Context, opts *graphql2.UserShiftsOptions) ([]oncall.Shift, error) {
	var userID string
	if opts.ID != "" {
		userID = opts.ID
	} else {
		userID = permission.UserID(ctx)
	}

	// get list of rotations user is on
	rIDs, err := (*App)(a).RotationStore.FindManyByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// get list of schedules user is on as a direct assignment, or indirectly from a rotation
	sIDs, err := (*App)(a).ScheduleStore.FindManyByAssignments(ctx, userID, rIDs)
	if err != nil {
		return nil, err
	}

	// calculate shifts for each schedule found
	var result []oncall.Shift
	for _, sID := range sIDs {
		shifts, err := (*App)(a).OnCallStore.ShiftsByUser(ctx, sID, opts.Start, opts.End, userID)
		if err != nil {
			return nil, err
		}
		result = append(result, shifts...)
	}

	return result, nil
}
