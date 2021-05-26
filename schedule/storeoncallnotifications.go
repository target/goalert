package schedule

import (
	"context"
	"database/sql"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const onCallNotificationRuleLimit = 50

func (store *Store) SetOnCallNotificationRules(ctx context.Context, tx *sql.Tx, scheduleID uuid.UUID, rules []OnCallNotificationRule) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	err = validate.Range("Rules", len(rules), 0, onCallNotificationRuleLimit)
	if err != nil {
		return err
	}

	ids := make([]bool, onCallNotificationRuleLimit)
	for i, r := range rules {
		if !r.ID.valid {
			continue
		}
		fieldName := fmt.Sprintf("Rules[%d].ID", i)
		err = validate.Range(fieldName, r.ID.id, 0, onCallNotificationRuleLimit)
		if err != nil {
			return err
		}
		if r.ID.scheduleID != scheduleID {
			return validation.NewFieldError(fieldName, "wrong schedule ID")
		}
		if ids[r.ID.id] {
			return validation.NewFieldError(fieldName, "duplicate ID value not allowed")
		}
		ids[r.ID.id] = true
	}
	nextID := func() int {
		for i, used := range ids {
			if used {
				continue
			}
			ids[i] = true
			return i
		}

		// should be impossible
		panic("could not find unused ID")
	}

	for i, r := range rules {
		if r.ID.valid {
			continue
		}

		rules[i].ID.scheduleID = scheduleID
		rules[i].ID.valid = true
		rules[i].ID.id = nextID()
	}

	return store.updateScheduleData(ctx, tx, scheduleID, func(data *Data) error {
		data.V1.OnCallNotificationRules = rules

		return nil
	})
}

func (store *Store) OnCallNotificationRules(ctx context.Context, tx *sql.Tx, id uuid.UUID) ([]OnCallNotificationRule, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	data, err := store.scheduleData(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return data.V1.OnCallNotificationRules, nil
}
