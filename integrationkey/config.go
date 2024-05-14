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
