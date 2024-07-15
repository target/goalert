package integrationkey

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const (
	MaxRules   = 100
	MaxActions = 10
	MaxParams  = 10
)

func ValidateUIKConfigV1(cfg gadb.UIKConfigV1) error {
	err := validate.Many(
		validate.Len("Rules", cfg.Rules, 0, MaxRules),
		validate.Len("DefaultActions", cfg.DefaultActions, 0, MaxActions),
	)
	if err != nil {
		return err
	}

	for i, r := range cfg.Rules {
		field := fmt.Sprintf("Rules[%d]", i)
		err := validate.Many(
			validate.Name(field+".Name", r.Name),
			validate.Text(field+".Description", r.Description, 0, 255), // these are arbitrary and will likely change as the feature is developed
			validate.Text(field+".ConditionExpr", r.ConditionExpr, 1, 1024),
			validate.Len(field+".Actions", r.Actions, 0, MaxActions),
		)
		if err != nil {
			return err
		}

		for j, a := range r.Actions {
			field := fmt.Sprintf("Rules[%d].Actions[%d]", i, j)
			err := validate.Many(
				validate.MapLen(field+".Dest.Args", a.Dest.Args, 0, MaxParams),
				validate.MapLen(field+".Params", a.Params, 0, MaxParams),
			)
			if err != nil {
				return err
			}
		}
	}

	for i, a := range cfg.DefaultActions {
		field := fmt.Sprintf("DefaultActions[%d]", i)
		err := validate.Many(
			validate.MapLen(field+".Dest.Args", a.Dest.Args, 0, MaxParams),
			validate.MapLen(field+".Params", a.Params, 0, MaxParams),
		)
		if err != nil {
			return err
		}
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	if len(data) > 64*1024 {
		return validation.NewFieldError("Config", "must be less than 64KiB in total")
	}

	return nil
}

func (s *Store) Config(ctx context.Context, db gadb.DBTX, keyID uuid.UUID) (*gadb.UIKConfigV1, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Service)
	if err != nil {
		return nil, err
	}

	cfg, err := gadb.New(db).IntKeyGetConfig(ctx, keyID)
	if errors.Is(err, sql.ErrNoRows) {
		return &gadb.UIKConfigV1{}, nil
	}
	if err != nil {
		return nil, err
	}

	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported config version: %d", cfg.Version)
	}

	return &cfg.V1, nil
}

func (s *Store) SetConfig(ctx context.Context, db gadb.DBTX, keyID uuid.UUID, cfg *gadb.UIKConfigV1) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	if cfg != nil {
		err := ValidateUIKConfigV1(*cfg)
		if err != nil {
			return err
		}
	}

	gdb := gadb.New(db)
	keyType, err := gdb.IntKeyGetType(ctx, keyID)
	if err != nil {
		return err
	}
	if keyType != gadb.EnumIntegrationKeysTypeUniversal {
		return validation.NewGenericError("config only supported for universal keys")
	}

	if cfg == nil {
		return gdb.IntKeyDeleteConfig(ctx, keyID)
	}

	// ensure all rule IDs are set, and all actions have a channel
	for i := range cfg.Rules {
		if cfg.Rules[i].ID == uuid.Nil {
			cfg.Rules[i].ID = uuid.New()
		}
		err := setActionChannels(ctx, gdb, cfg.Rules[i].Actions)
		if err != nil {
			return err
		}
	}
	err = setActionChannels(ctx, gdb, cfg.DefaultActions)
	if err != nil {
		return err
	}

	err = gadb.New(db).IntKeySetConfig(ctx, gadb.IntKeySetConfigParams{
		ID:     keyID,
		Config: gadb.UIKConfig{Version: 1, V1: *cfg},
	})
	if err != nil {
		return err
	}

	return nil
}

func setActionChannels(ctx context.Context, gdb *gadb.Queries, actions []gadb.UIKActionV1) error {
	for j, act := range actions {
		// We need to ensure the channel exists in the notification_channels table before we can use it.
		id, err := gdb.IntKeyEnsureChannel(ctx, gadb.IntKeyEnsureChannelParams{
			ID:   uuid.New(),
			Dest: gadb.NullDestV1{Valid: true, DestV1: act.Dest},
		})
		if err != nil {
			return err
		}
		actions[j].ChannelID = id
	}

	return nil
}
