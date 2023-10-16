package favorite

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"
)

// Store allows the lookup and management of Favorites.
type Store struct{}

// NewStore will create a new Store.
func NewStore(ctx context.Context) (*Store, error) { return &Store{}, nil }

// Set will store the target as a favorite of the given user. Must be authorized as System or the same user.
// It is safe to call multiple times.
func (s *Store) Set(ctx context.Context, tx gadb.DBTX, userID string, tgt assignment.Target) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.MatchUser(userID))
	if err != nil {
		return err
	}
	err = validate.Many(
		validate.UUID("TargetID", tgt.TargetID()),
		validate.UUID("UserID", userID),
		validate.OneOf("TargetType", tgt.TargetType(), assignment.TargetTypeService,
			assignment.TargetTypeSchedule, assignment.TargetTypeRotation, assignment.TargetTypeEscalationPolicy, assignment.TargetTypeUser),
	)
	if err != nil {
		return err
	}

	args := gadb.UserFavSetParams{UserID: uuid.MustParse(userID)}
	switch tgt.TargetType() {
	case assignment.TargetTypeService:
		args.TgtServiceID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	case assignment.TargetTypeSchedule:
		args.TgtScheduleID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	case assignment.TargetTypeRotation:
		args.TgtRotationID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	case assignment.TargetTypeEscalationPolicy:
		args.TgtEscalationPolicyID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	case assignment.TargetTypeUser:
		args.TgtUserID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	}

	err = gadb.New(tx).UserFavSet(ctx, args)
	if err != nil {
		return fmt.Errorf("set favorite: %w", err)
	}

	return nil
}

// Unset will remove the target as a favorite of the given user. Must be authorized as System or the same user.
// It is safe to call multiple times.
func (s *Store) Unset(ctx context.Context, tx gadb.DBTX, userID string, tgt assignment.Target) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.MatchUser(userID))
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("TargetID", tgt.TargetID()),
		validate.UUID("UserID", userID),
		validate.OneOf("TargetType", tgt.TargetType(), assignment.TargetTypeService,
			assignment.TargetTypeSchedule, assignment.TargetTypeRotation, assignment.TargetTypeEscalationPolicy, assignment.TargetTypeUser),
	)
	if err != nil {
		return err
	}

	args := gadb.UserFavUnsetParams{UserID: uuid.MustParse(userID)}
	switch tgt.TargetType() {
	case assignment.TargetTypeService:
		args.TgtServiceID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	case assignment.TargetTypeSchedule:
		args.TgtScheduleID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	case assignment.TargetTypeRotation:
		args.TgtRotationID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	case assignment.TargetTypeEscalationPolicy:
		args.TgtEscalationPolicyID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	case assignment.TargetTypeUser:
		args.TgtUserID = uuid.NullUUID{Valid: true, UUID: uuid.MustParse(tgt.TargetID())}
	}

	err = gadb.New(tx).UserFavUnset(ctx, args)
	if errors.Is(err, sql.ErrNoRows) {
		// ignoring since it is safe to unset favorite (with retries)
		err = nil
	}
	if err != nil {
		return fmt.Errorf("unset favorite: %w", err)
	}

	return nil
}

func (s *Store) FindAll(ctx context.Context, tx gadb.DBTX, userID string, filter []assignment.TargetType) ([]assignment.Target, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}

	err = validate.Many(
		validate.UUID("UserID", userID),
		validate.Range("Filter", len(filter), 0, 50),
	)
	if err != nil {
		return nil, err
	}

	args := gadb.UserFavFindAllParams{
		UserID:                  uuid.MustParse(userID),
		AllowServices:           len(filter) == 0 || slices.Contains(filter, assignment.TargetTypeService),
		AllowSchedules:          len(filter) == 0 || slices.Contains(filter, assignment.TargetTypeSchedule),
		AllowRotations:          len(filter) == 0 || slices.Contains(filter, assignment.TargetTypeRotation),
		AllowEscalationPolicies: len(filter) == 0 || slices.Contains(filter, assignment.TargetTypeEscalationPolicy),
		AllowUsers:              len(filter) == 0 || slices.Contains(filter, assignment.TargetTypeUser),
	}

	favs, err := gadb.New(tx).UserFavFindAll(ctx, args)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find all favorites: %w", err)
	}

	targets := make([]assignment.Target, 0, len(favs))
	for _, fav := range favs {
		switch {
		case fav.TgtServiceID.Valid:
			targets = append(targets, assignment.ServiceTarget(fav.TgtServiceID.UUID.String()))
		case fav.TgtScheduleID.Valid:
			targets = append(targets, assignment.ScheduleTarget(fav.TgtScheduleID.UUID.String()))
		case fav.TgtRotationID.Valid:
			targets = append(targets, assignment.RotationTarget(fav.TgtRotationID.UUID.String()))
		case fav.TgtEscalationPolicyID.Valid:
			targets = append(targets, assignment.EscalationPolicyTarget(fav.TgtEscalationPolicyID.UUID.String()))
		case fav.TgtUserID.Valid:
			targets = append(targets, assignment.UserTarget(fav.TgtUserID.UUID.String()))
		}
	}

	return targets, nil
}
