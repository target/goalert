package graphqlapp

import (
	"context"

	"github.com/target/goalert/apikey"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
)

func (a *Mutation) DeleteGQLAPIKey(ctx context.Context, input string) (bool, error) {
	id, err := parseUUID("ID", input)
	if err != nil {
		return false, err
	}

	err = a.APIKeyStore.DeleteAdminGraphQLKey(ctx, id)
	return err == nil, err
}

func (a *Mutation) CreateGQLAPIKey(ctx context.Context, input graphql2.CreateGQLAPIKeyInput) (*graphql2.CreatedGQLAPIKey, error) {
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
