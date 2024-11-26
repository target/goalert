package graphqlapp

import (
	context "context"
	"database/sql"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/search"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/timeutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

type Rotation App

func (a *App) Rotation() graphql2.RotationResolver { return (*Rotation)(a) }

func (q *Query) Rotation(ctx context.Context, id string) (*rotation.Rotation, error) {
	return (*App)(q).FindOneRotation(ctx, id)
}

func (m *Mutation) CreateRotation(ctx context.Context, input graphql2.CreateRotationInput) (result *rotation.Rotation, err error) {
	loc, err := util.LoadLocation(input.TimeZone)
	if err != nil {
		return nil, validation.NewFieldError("TimeZone", err.Error())
	}
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		rot := &rotation.Rotation{
			Name:  input.Name,
			Type:  input.Type,
			Start: input.Start.In(loc),
		}
		if input.Description != nil {
			rot.Description = *input.Description
		}
		if input.ShiftLength != nil {
			rot.ShiftLength = *input.ShiftLength
		}

		result, err = m.RotationStore.CreateRotationTx(ctx, tx, rot)
		if err != nil {
			return err
		}

		if input.Favorite != nil && *input.Favorite {
			err = m.FavoriteStore.Set(ctx, gadb.Compat(tx), permission.UserID(ctx), assignment.RotationTarget(result.ID))
			if err != nil {
				return err
			}
		}

		if input.UserIDs != nil {
			err := m.RotationStore.AddRotationUsersTx(ctx, tx, result.ID, input.UserIDs)
			if err != nil {
				return err
			}
		}
		return err
	})

	return result, err
}

func (r *Rotation) TimeZone(ctx context.Context, rot *rotation.Rotation) (string, error) {
	return rot.Start.Location().String(), nil
}

func (r *Rotation) IsFavorite(ctx context.Context, rot *rotation.Rotation) (bool, error) {
	return rot.IsUserFavorite(), nil
}

func (r *Rotation) NextHandoffTimes(ctx context.Context, rot *rotation.Rotation, num *int) ([]time.Time, error) {
	var n int
	if num != nil {
		n = *num
	} else {
		count, err := r.RotationStore.FindParticipantCount(ctx, rot.ID)
		if err != nil {
			return nil, errors.Wrap(err, "retrieving participant count")
		}
		if count > 50 {
			// setting to max limit for validation
			n = 50
		} else {
			n = count
		}
	}

	err := validate.Range("num", n, 0, 50)
	if err != nil {
		return nil, err
	}

	s, err := r.RotationStore.State(ctx, rot.ID)
	if errors.Is(err, rotation.ErrNoState) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	result := make([]time.Time, n)
	t := s.ShiftStart
	for i := range result {
		t = rot.EndTime(t)
		result[i] = t
	}

	return result, nil
}

func (r *Rotation) UserIDs(ctx context.Context, rot *rotation.Rotation) ([]string, error) {
	parts, err := r.RotationStore.FindAllParticipants(ctx, rot.ID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(parts))
	for i, p := range parts {
		ids[i] = p.Target.TargetID()
	}

	return ids, nil
}

func (r *Rotation) Users(ctx context.Context, rot *rotation.Rotation) ([]user.User, error) {
	userIDs, err := r.UserIDs(ctx, rot)
	if err != nil {
		return nil, err
	}

	users := make([]user.User, len(userIDs))
	errCh := make(chan error, len(userIDs))
	for i := range userIDs {
		// TODO: does this need to be bounded?
		// The max number = the max number of unique users in the current rotation,
		// which can be bounded by the config_limit participants_per_rotation (but isn't by default).
		go func(idx int) {
			u, err := (*App)(r).FindOneUser(ctx, userIDs[idx])
			if err == nil {
				users[idx] = *u
			}
			errCh <- err
		}(i)
	}

	for range userIDs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errCh:
			if err != nil {
				return nil, err
			}
		}
	}

	return users, nil
}

func (r *Rotation) ActiveUserIndex(ctx context.Context, obj *rotation.Rotation) (int, error) {
	s, err := r.RotationStore.State(ctx, obj.ID)
	if errors.Is(err, rotation.ErrNoState) {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	return s.Position, err
}

func (q *Query) Rotations(ctx context.Context, opts *graphql2.RotationSearchOptions) (conn *graphql2.RotationConnection, err error) {
	if opts == nil {
		opts = &graphql2.RotationSearchOptions{}
	}

	var searchOpts rotation.SearchOptions
	searchOpts.FavoritesUserID = permission.UserID(ctx)
	if opts.Search != nil {
		searchOpts.Search = *opts.Search
	}
	if opts.FavoritesFirst != nil {
		searchOpts.FavoritesFirst = *opts.FavoritesFirst
	}

	if opts.FavoritesOnly != nil {
		searchOpts.FavoritesOnly = *opts.FavoritesOnly
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
	rots, err := q.RotationStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}
	conn = new(graphql2.RotationConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(rots) == searchOpts.Limit {
		rots = rots[:len(rots)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(rots) > 0 {
		last := rots[len(rots)-1]
		searchOpts.After.IsFavorite = last.IsUserFavorite()
		searchOpts.After.Name = last.Name

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = rots
	return conn, err
}

func (m *Mutation) updateRotationParticipants(ctx context.Context, tx *sql.Tx, rotationID string, userIDs []string, updateActive bool) (err error) {
	// Get current participants
	currentParticipants, err := m.RotationStore.FindAllParticipantsTx(ctx, tx, rotationID)
	if err != nil {
		return err
	}

	var participantIDsToRemove []string

	for i, c := range currentParticipants {
		if i >= len(userIDs) {
			participantIDsToRemove = append(participantIDsToRemove, c.ID)
			continue
		}

		if c.Target.TargetID() == userIDs[i] {
			// nothing to update
			continue
		}

		// Update
		err = m.RotationStore.UpdateParticipantUserIDTx(ctx, tx, c.ID, userIDs[i])
		if err != nil {
			return err
		}
	}

	if len(userIDs) > len(currentParticipants) {
		// Add users
		err = m.RotationStore.AddRotationUsersTx(ctx, tx, rotationID, userIDs[len(currentParticipants):])
		if err != nil {
			return err
		}
	}

	if len(participantIDsToRemove) == 0 {
		return nil
	}

	if len(userIDs) == 0 {
		// Delete rotation state if all users are going to be deleted as per new input
		err = m.RotationStore.DeleteStateTx(ctx, tx, rotationID)
		if err != nil {
			return err
		}
	} else if updateActive {
		// get current active participant
		s, err := m.RotationStore.StateTx(ctx, tx, rotationID)
		if errors.Is(err, rotation.ErrNoState) {
			return nil
		}
		if err != nil {
			return err
		}

		// if currently active user is going to be deleted
		// then set to first user before we actually delete any users
		if s.Position >= len(userIDs) {
			err = m.RotationStore.SetActiveIndexTx(ctx, tx, rotationID, 0)
			if err != nil {
				return err
			}
		}
	}

	err = m.RotationStore.DeleteRotationParticipantsTx(ctx, tx, participantIDsToRemove)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mutation) UpdateRotation(ctx context.Context, input graphql2.UpdateRotationInput) (res bool, err error) {
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		result, err := m.RotationStore.FindRotationForUpdateTx(ctx, tx, input.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return validation.NewFieldError("id", "Rotation not found")
		}
		if err != nil {
			return err
		}
		var update bool
		if input.Name != nil {
			update = true
			result.Name = *input.Name
		}
		if input.Description != nil {
			update = true
			result.Description = *input.Description
		}
		if input.Start != nil {
			update = true
			result.Start = *input.Start
		}
		if input.Type != nil {
			update = true
			result.Type = *input.Type
		}
		if input.ShiftLength != nil {
			update = true
			result.ShiftLength = *input.ShiftLength
		}

		if input.TimeZone != nil {
			update = true
			loc, err := util.LoadLocation(*input.TimeZone)
			if err != nil {
				return validation.NewFieldError("TimeZone", "invalid TimeZone: "+err.Error())
			}
			result.Start = result.Start.In(loc)
		}

		if update {
			err = m.RotationStore.UpdateRotationTx(ctx, tx, result)
			if err != nil {
				return err
			}
		}

		if input.UserIDs != nil {
			err = m.updateRotationParticipants(ctx, tx, input.ID, input.UserIDs, input.ActiveUserIndex == nil)
			if err != nil {
				return err
			}
		}

		// Update active participant (in rotation state) if specified by input
		// This should be applicable regardless of whether or not 'UserIDs' as an input has been specified.
		if input.ActiveUserIndex != nil {
			err = m.RotationStore.SetActiveIndexTx(ctx, tx, input.ID, *input.ActiveUserIndex)
			if err != nil {
				return err
			}
		}

		return err
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *Query) CalcRotationHandoffTimes(ctx context.Context, input *graphql2.CalcRotationHandoffTimesInput) ([]time.Time, error) {
	err := validate.Range("count", input.Count, 0, 20)
	if err != nil {
		return nil, err
	}

	loc, err := util.LoadLocation(input.TimeZone)
	if err != nil {
		return nil, validation.NewFieldError("timeZone", err.Error())
	}

	if input.ShiftLength != nil && input.ShiftLengthHours != nil {
		return nil, validation.NewFieldError("shiftLength", "only one of (shiftLength, shiftLengthHours) is allowed")
	}

	rot := rotation.Rotation{
		Start: input.Handoff.In(loc),
	}
	switch {
	case input.ShiftLength != nil:
		err = setRotationShiftFromISO(&rot, input.ShiftLength)
		if err != nil {
			return nil, err
		}
	case input.ShiftLengthHours != nil:
		err = validate.Range("hours", *input.ShiftLengthHours, 0, 99999)
		if err != nil {
			return nil, err
		}
		rot.Type = rotation.TypeHourly
		rot.ShiftLength = *input.ShiftLengthHours
	default:
		return nil, validation.NewFieldError("shiftLength", "must be specified")
	}

	t := time.Now()
	if input.From != nil {
		t = input.From.In(loc)
	}

	var result []time.Time
	for len(result) < input.Count {
		t = rot.EndTime(t)
		result = append(result, t)
	}

	return result, nil
}

// getRotationFromISO determines the rotation type based on the given ISODuration. An error is given if the unsupported year field or multiple non-zero fields are given.
func setRotationShiftFromISO(rot *rotation.Rotation, dur *timeutil.ISODuration) error {
	// validate only one time field (year, month, days, timepart) is non-zero
	nonZeroFields := 0

	if dur.YearPart > 0 {
		// These validation errors are only possible from direct api calls,
		// thus using ISO standard terminology "designator" to match the spec.
		return validation.NewFieldError("shiftLength", "year designator not allowed")
	}

	if dur.MonthPart > 0 {
		rot.Type = rotation.TypeMonthly
		rot.ShiftLength = dur.MonthPart
		nonZeroFields++
	}
	if dur.WeekPart > 0 {
		rot.Type = rotation.TypeWeekly
		rot.ShiftLength = dur.WeekPart
		nonZeroFields++
	}
	if dur.DayPart > 0 {
		rot.Type = rotation.TypeDaily
		rot.ShiftLength = dur.DayPart
		nonZeroFields++
	}
	if dur.HourPart > 0 {
		rot.Type = rotation.TypeHourly
		rot.ShiftLength = dur.HourPart
		nonZeroFields++
	}

	if nonZeroFields == 0 {
		return validation.NewFieldError("shiftLength", "must not be zero")
	}
	if nonZeroFields > 1 {
		// Same as above, this error is only possible from direct api calls.
		return validation.NewFieldError("shiftLength", "only one of (M, W, D, H) is allowed")
	}

	return nil
}
