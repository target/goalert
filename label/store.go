package label

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"

	"github.com/pkg/errors"
)

// Store allows the lookup and management of Labels.
type Store struct {
	db *sql.DB
}

// NewStore will Set a DB backend from a sql.DB. An error will be returned if statements fail to prepare.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) { return &Store{db: db}, nil }

func (l *Label) tgtEP() uuid.NullUUID {
	if l.Target.TargetType() != assignment.TargetTypeEscalationPolicy {
		return uuid.NullUUID{}
	}

	return uuid.NullUUID{UUID: uuid.MustParse(l.Target.TargetID()), Valid: true}
}

func (l *Label) tgtSvc() uuid.NullUUID {
	if l.Target.TargetType() != assignment.TargetTypeService {
		return uuid.NullUUID{}
	}

	return uuid.NullUUID{UUID: uuid.MustParse(l.Target.TargetID()), Valid: true}
}

func (l *Label) tgtSched() uuid.NullUUID {
	if l.Target.TargetType() != assignment.TargetTypeSchedule {
		return uuid.NullUUID{}
	}

	return uuid.NullUUID{UUID: uuid.MustParse(l.Target.TargetID()), Valid: true}
}

func (l *Label) tgtRot() uuid.NullUUID {
	if l.Target.TargetType() != assignment.TargetTypeRotation {
		return uuid.NullUUID{}
	}

	return uuid.NullUUID{UUID: uuid.MustParse(l.Target.TargetID()), Valid: true}
}

// SetTx will set a label for the service. It can be used to set the key-value pair for the label,
// delete a label or update the value given the label's key.
func (s *Store) SetTx(ctx context.Context, db gadb.DBTX, label *Label) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return err
	}

	n, err := label.Normalize()
	if err != nil {
		return err
	}

	if n.Value == "" { // delete if value is empty
		err = gadb.New(db).LabelDeleteKeyByTarget(ctx, gadb.LabelDeleteKeyByTargetParams{
			Key:        label.Key,
			ServiceID:  label.tgtSvc(),
			ScheduleID: label.tgtSched(),
			RotationID: label.tgtRot(),
			EpID:       label.tgtEP(),
		})
		if err != nil {
			return fmt.Errorf("delete label: %w", err)
		}

		return nil
	}

	err = gadb.New(db).LabelSetByTarget(ctx, gadb.LabelSetByTargetParams{
		Key:        label.Key,
		Value:      label.Value,
		ServiceID:  label.tgtSvc(),
		ScheduleID: label.tgtSched(),
		RotationID: label.tgtRot(),
		EpID:       label.tgtEP(),
	})
	if err != nil {
		return fmt.Errorf("set label: %w", err)
	}

	return nil
}

// FindAllByTarget finds all labels for a particular target. It returns all key-value pairs.
func (s *Store) FindAllByTarget(ctx context.Context, db gadb.DBTX, t assignment.Target) ([]Label, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	label := Label{Target: t}
	rows, err := gadb.New(db).LabelFindAllByTarget(ctx, gadb.LabelFindAllByTargetParams{
		ServiceID:  label.tgtSvc(),
		ScheduleID: label.tgtSched(),
		RotationID: label.tgtRot(),
		EpID:       label.tgtEP(),
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find all labels by service: %w", err)
	}

	labels := make([]Label, len(rows))
	for i, l := range rows {
		labels[i].Key = l.Key
		labels[i].Value = l.Value
		labels[i].Target = t
	}

	return labels, nil
}

func (s *Store) UniqueKeysTx(ctx context.Context, db gadb.DBTX) ([]string, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	return gadb.New(db).LabelUniqueKeys(ctx)
}
