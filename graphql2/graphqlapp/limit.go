package graphqlapp

import (
	"context"
	"database/sql"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/limit"
	"github.com/target/goalert/permission"
)

func (q *Query) SystemLimits(ctx context.Context) ([]graphql2.SystemLimit, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return nil, err
	}
	limits, err := q.LimitStore.All(ctx)
	if err != nil {
		return nil, err
	}
	return graphql2.MapLimitValues(limits), nil
}
func (m *Mutation) SetSystemLimits(ctx context.Context, input []graphql2.SystemLimitInput) (bool, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin)
	if err != nil {
		return false, err
	}
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		l := limit.Limits{}
		l, err := graphql2.ApplyLimitValues(l, input)

		for id, max := range l {
			err = m.LimitStore.UpdateLimitsTx(ctx, tx, string(id), max)
		}
		return err
	})
	return err == nil, err
}
