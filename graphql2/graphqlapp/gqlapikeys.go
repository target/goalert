package graphqlapp

import (
	"context"

	"github.com/target/goalert/apikey"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
)

func (a *Mutation) DeleteGQLAPIKey(ctx context.Context, input string) (bool, error) {
	if !expflag.ContextHas(ctx, expflag.GQLAPIKey) {
		return false, validation.NewGenericError("experimental flag not enabled")
	}
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
