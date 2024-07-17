package escalation

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

func (s *Store) AddStepActionTx(ctx context.Context, tx *sql.Tx, stepID uuid.UUID, dest gadb.DestV1) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	typeInfo, err := s.reg.TypeInfo(ctx, dest.Type)
	if err != nil {
		return err
	}
	if !typeInfo.IsEPTarget() {
		return validation.NewGenericError("not a valid escalation policy destination")
	}

	var userID, scheduleID, rotationID, channelID uuid.NullUUID
	switch dest.Type {
	case user.DestTypeUser:
		id, err := validate.ParseUUID("ID", dest.Arg(user.FieldUserID))
		if err != nil {
			return err
		}
		userID = uuid.NullUUID{UUID: id, Valid: true}
	case schedule.DestTypeSchedule:
		id, err := validate.ParseUUID("ID", dest.Arg(schedule.FieldScheduleID))
		if err != nil {
			return err
		}
		scheduleID = uuid.NullUUID{UUID: id, Valid: true}
	case rotation.DestTypeRotation:
		id, err := validate.ParseUUID("ID", dest.Arg(rotation.FieldRotationID))
		if err != nil {
			return err
		}
		rotationID = uuid.NullUUID{UUID: id, Valid: true}
	default:
		id, err := s.ncStore.MapDestToID(ctx, tx, dest)
		if err != nil {
			return err
		}
		channelID = uuid.NullUUID{UUID: id, Valid: true}
	}

	return gadb.New(tx).EPStepActionsAddAction(ctx, gadb.EPStepActionsAddActionParams{
		EscalationPolicyStepID: stepID,
		UserID:                 userID,
		ScheduleID:             scheduleID,
		RotationID:             rotationID,
		ChannelID:              channelID,
	})
}

func (s *Store) DeleteStepActionTx(ctx context.Context, tx *sql.Tx, stepID uuid.UUID, dest gadb.DestV1) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	var userID, scheduleID, rotationID, channelID uuid.NullUUID
	switch dest.Type {
	case user.DestTypeUser:
		id, err := validate.ParseUUID("ID", dest.Arg(user.FieldUserID))
		if err != nil {
			return err
		}
		userID = uuid.NullUUID{UUID: id, Valid: true}
	case schedule.DestTypeSchedule:
		id, err := validate.ParseUUID("ID", dest.Arg(schedule.FieldScheduleID))
		if err != nil {
			return err
		}
		scheduleID = uuid.NullUUID{UUID: id, Valid: true}
	case rotation.DestTypeRotation:
		id, err := validate.ParseUUID("ID", dest.Arg(rotation.FieldRotationID))
		if err != nil {
			return err
		}
		rotationID = uuid.NullUUID{UUID: id, Valid: true}
	default:
		id, err := s.ncStore.LookupDestID(ctx, tx, dest)
		if err != nil {
			return err
		}
		channelID = uuid.NullUUID{UUID: id, Valid: true}
	}

	return gadb.New(tx).EPStepActionsDeleteAction(ctx, gadb.EPStepActionsDeleteActionParams{
		EscalationPolicyStepID: stepID,
		UserID:                 userID,
		ScheduleID:             scheduleID,
		RotationID:             rotationID,
		ChannelID:              channelID,
	})
}

func (s *Store) FindAllStepActionsTx(ctx context.Context, tx *sql.Tx, stepID uuid.UUID) ([]gadb.DestV1, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	actions, err := gadb.New(tx).EPStepActionsByStepId(ctx, stepID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	result := make([]gadb.DestV1, 0, len(actions))
	for _, a := range actions {
		switch {
		case a.UserID.Valid:
			result = append(result, user.DestFromID(a.UserID.UUID.String()))
		case a.ScheduleID.Valid:
			result = append(result, schedule.DestFromID(a.ScheduleID.UUID.String()))
		case a.RotationID.Valid:
			result = append(result, rotation.DestFromID(a.RotationID.UUID.String()))
		case a.Dest.Valid:
			result = append(result, a.Dest.DestV1)
		}
	}

	return result, nil
}
