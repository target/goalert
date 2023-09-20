package graphqlapp

import (
	"context"
	"database/sql"
	"time"

	"github.com/target/goalert/apikey"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"
)

type GQLAPIKey App

func (a *App) GQLAPIKey() graphql2.GQLAPIKeyResolver { return (*GQLAPIKey)(a) }

func (a *GQLAPIKey) CreatedBy(ctx context.Context, obj *graphql2.GQLAPIKey) (*user.User, error) {
	if obj.CreatedBy == nil {
		return nil, nil
	}

	return (*App)(a).FindOneUser(ctx, obj.CreatedBy.ID)
}

func (a *GQLAPIKey) UpdatedBy(ctx context.Context, obj *graphql2.GQLAPIKey) (*user.User, error) {
	if obj.UpdatedBy == nil {
		return nil, nil
	}

	return (*App)(a).FindOneUser(ctx, obj.UpdatedBy.ID)
}

func (q *Query) GqlAPIKeys(ctx context.Context) ([]graphql2.GQLAPIKey, error) {
	if !expflag.ContextHas(ctx, expflag.GQLAPIKey) {
		return nil, validation.NewGenericError("experimental flag not enabled")
	}
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	keys, err := q.APIKeyStore.FindAllAdminGraphQLKeys(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]graphql2.GQLAPIKey, len(keys))
	for i, k := range keys {
		res[i] = graphql2.GQLAPIKey{
			ID:            k.ID.String(),
			Name:          k.Name,
			Description:   k.Description,
			CreatedAt:     k.CreatedAt,
			UpdatedAt:     k.UpdatedAt,
			ExpiresAt:     k.ExpiresAt,
			AllowedFields: k.AllowedFields,
		}

		if k.CreatedBy != nil {
			res[i].CreatedBy = &user.User{ID: k.CreatedBy.String()}
		}
		if k.UpdatedBy != nil {
			res[i].UpdatedBy = &user.User{ID: k.UpdatedBy.String()}
		}

		if k.LastUsed != nil {
			res[i].LastUsed = &graphql2.GQLAPIKeyUsage{
				Time: k.LastUsed.Time,
				Ua:   k.LastUsed.UserAgent,
				IP:   k.LastUsed.IP,
			}
		}
	}

	return res, nil
}

func nullTimeToPointer(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

func (a *Mutation) UpdateGQLAPIKey(ctx context.Context, input graphql2.UpdateGQLAPIKeyInput) (bool, error) {
	if !expflag.ContextHas(ctx, expflag.GQLAPIKey) {
		return false, validation.NewGenericError("experimental flag not enabled")
	}
	id, err := parseUUID("ID", input.ID)
	if err != nil {
		return false, err
	}

	err = a.APIKeyStore.UpdateAdminGraphQLKey(ctx, id, input.Name, input.Description)
	return err == nil, err
}

func (a *Mutation) DeleteGQLAPIKey(ctx context.Context, input string) (bool, error) {
	id, err := parseUUID("ID", input)
	if err != nil {
		return false, err
	}

	err = a.APIKeyStore.DeleteAdminGraphQLKey(ctx, id)
	return err == nil, err
}

func (a *Mutation) CreateGQLAPIKey(ctx context.Context, input graphql2.CreateGQLAPIKeyInput) (*graphql2.CreatedGQLAPIKey, error) {
	if !expflag.ContextHas(ctx, expflag.GQLAPIKey) {
		return nil, validation.NewGenericError("experimental flag not enabled")
	}

	id, tok, err := a.APIKeyStore.CreateAdminGraphQLKey(ctx, apikey.NewAdminGQLKeyOpts{
		Name:    input.Name,
		Desc:    input.Description,
		Expires: input.ExpiresAt,
		Fields:  input.AllowedFields,
		Role:    permission.Role(input.Role),
	})
	if err != nil {
		return nil, err
	}

	return &graphql2.CreatedGQLAPIKey{
		ID:    id.String(),
		Token: tok,
	}, nil
}
