package override

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"
)

// SearchOptions allow filtering and paginating the list of rotations.
type SearchOptions struct {
	After SearchCursor `json:"a,omitempty"`
	Limit int          `json:"-"`

	// Omit specifies a list of override IDs to exclude from the results.
	Omit []string

	ScheduleID string `json:"d,omitempty"`

	AddUserIDs    []string `json:"u,omitempty"`
	RemoveUserIDs []string `json:"r,omitempty"`
	AnyUserIDs    []string `json:"n,omitempty"`

	Start time.Time `json:"t,omitempty"`
	End   time.Time `json:"e,omitempty"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	ID string `json:"i,omitempty"`
}

func (s *Store) Search(ctx context.Context, db gadb.DBTX, opts *SearchOptions) ([]UserOverride, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &SearchOptions{}
	}

	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	var arg gadb.OverrideSearchParams
	arg.AddUserID, err = validate.ParseManyUUID("AddUserIDs", opts.AddUserIDs, 10)
	if err != nil {
		return nil, err
	}
	arg.RemoveUserID, err = validate.ParseManyUUID("RemoveUserIDs", opts.RemoveUserIDs, 10)
	if err != nil {
		return nil, err
	}
	arg.AnyUserID, err = validate.ParseManyUUID("AnyUserIDs", opts.AnyUserIDs, 10)
	if err != nil {
		return nil, err
	}
	arg.Omit, err = validate.ParseManyUUID("Omit", opts.Omit, 50)
	if err != nil {
		return nil, err
	}
	if opts.ScheduleID != "" {
		id, err := validate.ParseUUID("ScheduleID", opts.ScheduleID)
		if err != nil {
			return nil, err
		}
		arg.ScheduleID = uuid.NullUUID{UUID: id, Valid: true}
	}
	if !opts.Start.IsZero() {
		arg.SearchStart = sql.NullTime{Time: opts.Start, Valid: true}
	}
	if !opts.End.IsZero() {
		arg.SearchEnd = sql.NullTime{Time: opts.End, Valid: true}
	}
	if opts.After.ID != "" {
		id, err := validate.ParseUUID("After.ID", opts.After.ID)
		if err != nil {
			return nil, err
		}
		arg.AfterID = uuid.NullUUID{UUID: id, Valid: true}
	}

	rows, err := gadb.New(db).OverrideSearch(ctx, arg)
	if err != nil {
		return nil, err
	}
	if len(rows) > opts.Limit && opts.Limit > 0 {
		rows = rows[:opts.Limit]
	}

	result := make([]UserOverride, len(rows))
	for i, r := range rows {
		var add, rem string
		if r.AddUserID.Valid {
			add = r.AddUserID.UUID.String()
		}
		if r.RemoveUserID.Valid {
			rem = r.RemoveUserID.UUID.String()
		}

		result[i] = UserOverride{
			ID:           r.ID.String(),
			Start:        r.StartTime,
			End:          r.EndTime,
			AddUserID:    add,
			RemoveUserID: rem,
			Target:       assignment.ScheduleTarget(r.TgtScheduleID.String()),
		}
	}

	return result, nil
}
