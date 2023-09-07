package graphqlapp

import (
	"context"
	"fmt"
	"sort"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/validation"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

func (q *Query) GqlAPIKeys(ctx context.Context) ([]graphql2.GQLAPIKey, error) {
	return nil, nil
}

func (q *Query) ListGQLFields(ctx context.Context, query *string) ([]string, error) {
	if query == nil || *query == "" {
		sch, err := parser.ParseSchema(&ast.Source{Input: graphql2.Schema()})
		if err != nil {
			return nil, fmt.Errorf("parse schema: %w", err)
		}

		var fields []string
		for _, typ := range sch.Definitions {
			if typ.Kind != ast.Object {
				continue
			}
			for _, f := range typ.Fields {
				fields = append(fields, typ.Name+"."+f.Name)
			}
		}
		sort.Strings(fields)
		return fields, nil
	}

	sch, err := gqlparser.LoadSchema(&ast.Source{Input: graphql2.Schema()})
	if err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	qDoc, qErr := gqlparser.LoadQuery(sch, *query)
	if len(qErr) > 0 {
		return nil, validation.NewFieldError("Query", qErr.Error())
	}

	var fields []string
	var e validator.Events
	e.OnField(func(w *validator.Walker, field *ast.Field) {
		fields = append(fields, field.ObjectDefinition.Name+"."+field.Name)
	})
	validator.Walk(sch, qDoc, &e)

	sort.Strings(fields)
	return fields, nil
}

func (a *Mutation) UpdateGQLAPIKey(ctx context.Context, input graphql2.UpdateGQLAPIKeyInput) (bool, error) {
	return false, nil
}
func (a *Mutation) DeleteGQLAPIKey(ctx context.Context, input string) (bool, error) {
	return false, nil
}
func (a *Mutation) CreateGQLAPIKey(ctx context.Context, input graphql2.CreateGQLAPIKeyInput) (*graphql2.GQLAPIKey, error) {
	return nil, nil
	// _, err := gqlauth.NewQuery(input.Query)
	// if err != nil {
	// 	return nil, validation.NewFieldError("Query", err.Error())
	// }

	// key, err := a.APIKeyStore.CreateAdminGraphQLKey(ctx, input.Name, input.Query, input.ExpiresAt)
	// if err != nil {
	// 	return nil, err
	// }

	// return &graphql2.GQLAPIKey{
	// 	ID:        key.ID.String(),
	// 	Name:      key.Name,
	// 	ExpiresAt: key.ExpiresAt,
	// 	Token:     &key.Token,
	// }, nil
}
