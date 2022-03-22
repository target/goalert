package graphqlapp

import (
	"context"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/validation"
)

func (a *Query) SwoStatus(ctx context.Context) (*graphql2.SWOStatus, error) {
	if a.SWO == nil {
		return nil, validation.NewGenericError("not in SWO mode")
	}

	s := a.SWO.Status()
	var nodes []graphql2.SWONode
	for _, n := range s.Nodes {
		nodes = append(nodes, graphql2.SWONode{
			ID:       n.ID.String(),
			OldValid: n.OldValid,
			NewValid: n.NewValid,
			CanExec:  n.CanExec,
			Status:   n.Status,
		})
	}

	return &graphql2.SWOStatus{
		IsIdle:  s.IsIdle,
		IsDone:  s.IsDone,
		Details: s.Details,
		Nodes:   nodes,
	}, nil
}
