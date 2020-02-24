package graphqlapp

import (
	"context"
	"database/sql"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/limit"
)

func (q *Query) SystemLimits(ctx context.Context) ([]graphql2.SystemLimit, error) {
	limits, err := q.LimitStore.All(ctx)
	if err != nil {
		return nil, err
	}
	return graphql2.MapLimitValues(limits), nil
}
func (m *Mutation) SetSystemLimits(ctx context.Context, input []graphql2.SystemLimitInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		l := limit.Limits{}
		l, err := graphql2.ApplyLimitValues(l, input)
		if err != nil {
			return err
		}

		for id, max := range l {
			err = m.LimitStore.UpdateLimitsTx(ctx, tx, string(id), max)
			if err != nil {
				return err
			}
		}
		return err
	})
	return err == nil, err
}
