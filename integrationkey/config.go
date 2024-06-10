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

type dbConfig struct {
	Version int // should be 1
	V1      Config
}

// Config stores the configuration for an integration key for how to handle incoming requests.
type Config struct {
	Rules []Rule

	// DefaultActions are the actions to take if no rules match.
	DefaultActions []Action

	StopAfterFirstMatchingRule bool
}

// A Rule is a set of conditions and actions to take if those conditions are met.
type Rule struct {
	ID            uuid.UUID
	Name          string
	Description   string
	ConditionExpr string
	Actions       []Action

	ContinueAfterMatch bool
}

// An Action is a single action to take if a rule matches.
type Action struct {
	// Type is the type of action to perform, like slack, email, or alert.
	Type string

	// StaticParams are parameters that are always the same for this action (e.g., the channel ID).
	StaticParams map[string]string

	// DynamicParams are parameters that are determined at runtime (e.g., the message to send).
	// The keys are the parameter names, and the values are the Expr expression strings.
	DynamicParams map[string]string
}

func (cfg Config) Validate() error {
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
				validate.MapLen(field+".StaticParams", a.StaticParams, 0, MaxParams),
				validate.MapLen(field+".DynamicParams", a.DynamicParams, 0, MaxParams),
			)
			if err != nil {
				return err
			}
		}
	}

	for i, a := range cfg.DefaultActions {
		field := fmt.Sprintf("DefaultActions[%d]", i)
		err := validate.Many(
			validate.MapLen(field+".StaticParams", a.StaticParams, 0, MaxParams),
			validate.MapLen(field+".DynamicParams", a.DynamicParams, 0, MaxParams),
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

func (s *Store) Config(ctx context.Context, db gadb.DBTX, keyID uuid.UUID) (*Config, error) {
	err := permission.LimitCheckAny(ctx, permission.User, permission.Service)
	if err != nil {
		return nil, err
	}

	data, err := gadb.New(db).IntKeyGetConfig(ctx, keyID)
	if errors.Is(err, sql.ErrNoRows) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg dbConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported config version: %d", cfg.Version)
	}

	return &cfg.V1, nil
}

func (s *Store) SetConfig(ctx context.Context, db gadb.DBTX, keyID uuid.UUID, cfg *Config) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	if cfg != nil {
		err := cfg.Validate()
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

	// ensure all rule IDs are set
	for i := range cfg.Rules {
		if cfg.Rules[i].ID == uuid.Nil {
			cfg.Rules[i].ID = uuid.New()
		}
	}

	data, err := json.Marshal(dbConfig{Version: 1, V1: *cfg})
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = gadb.New(db).IntKeySetConfig(ctx, gadb.IntKeySetConfigParams{
		ID:     keyID,
		Config: data,
	})
	if err != nil {
		return err
	}

	return nil
}
