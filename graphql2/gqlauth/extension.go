package gqlauth

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/target/goalert/permission"
)

func (q *Query) ExtensionName() string {
	return "gqlauth"
}

func (q *Query) Validate(schema graphql.ExecutableSchema) error {
	return nil
}
func (q *Query) InterceptField(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	f := graphql.GetFieldContext(ctx)
	err = q.ValidateField(f.Field.Field)
	if err != nil {
		return nil, permission.NewAccessDenied(err.Error())
	}

	return next(ctx)
}
