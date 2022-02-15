package calsub

import (
	"context"
	"errors"

	"github.com/target/goalert/auth/authtoken"
	"github.com/target/goalert/keyring"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Store allows the lookup and management of calendar subscriptions
type Store struct {
	keys keyring.Keyring
	oc   oncall.Store
}

// NewStore will create a new Store with the given parameters.
func NewStore(ctx context.Context, apiKeyring keyring.Keyring, oc oncall.Store) (*Store, error) {
	return &Store{
		keys: apiKeyring,
		oc:   oc,
	}, nil
}

// Authorize will return an authorized context associated with the given token. If the token is invalid
// or otherwise can not be authenticated, an error is returned.
func Authorize(ctx context.Context, tok authtoken.Token) (context.Context, error) {
	if tok.Type != authtoken.TypeCalSub {
		return ctx, validation.NewFieldError("token", "invalid type")
	}

	sCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	sCtx = permission.SystemContext(sCtx, "CalSubAuthorize")

	var cs Subscription
	err := sqlutil.FromContext(ctx).
		WithContext(sCtx).
		Model(&cs).
		Where("not disabled").
		Where("id = ?", tok.ID).
		Where("date_trunc('second', created_at) = ?", tok.CreatedAt).
		Clauses(clause.Returning{}).
		Update("last_access", gorm.Expr("now()")).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx, validation.NewFieldError("sub", "invalid")
	}
	if err != nil {
		return ctx, err
	}

	return permission.UserSourceContext(ctx, cs.UserID, permission.RoleUser, &permission.SourceInfo{
		Type: permission.SourceTypeCalendarSubscription,
		ID:   tok.ID.String(),
	}), nil
}

func (s *Store) SignToken(ctx context.Context, cs *Subscription) (string, error) {
	if cs.token == nil {
		return "", nil
	}

	return cs.token.Encode(s.keys.Sign)
}
