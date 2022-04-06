package graphqlapp

import (
	"context"
	"strings"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/validation"
)

func (m *Mutation) SwoAction(ctx context.Context, action graphql2.SWOAction) (bool, error) {
	if m.SWO == nil {
		return false, validation.NewGenericError("not in SWO mode")
	}

	var err error
	switch action {
	case graphql2.SWOActionPing:
		err = m.SWO.SendPing(ctx)
	case graphql2.SWOActionReset:
		err = m.SWO.SendReset(ctx)
	case graphql2.SWOActionExecute:
		err = m.SWO.SendExecute(ctx)
	default:
		return false, validation.NewGenericError("invalid SWO action")
	}

	return err == nil, err
}

func (a *Query) SwoStatus(ctx context.Context) (*graphql2.SWOStatus, error) {
	if a.SWO == nil {
		return nil, validation.NewGenericError("not in SWO mode")
	}

	s := a.SWO.Status()
	var nodes []graphql2.SWONode
	var prog string
	for _, n := range s.Nodes {
		var tasks []string
		for _, t := range n.Tasks {
			tasks = append(tasks, t.Name)
			if t.Name == "reset-db" || t.Name == "exec" {
				prog = t.Status
			}
		}

		nodes = append(nodes, graphql2.SWONode{
			ID:       n.ID.String(),
			OldValid: n.OldDBValid,
			NewValid: n.NewDBValid,
			IsLeader: n.IsLeader,
			CanExec:  n.CanExec,
			Status:   strings.Join(tasks, ","),
		})
	}

	status := string(s.State)
	if prog != "" {
		status += ": " + prog
	}

	return &graphql2.SWOStatus{
		IsIdle:      s.State == "idle",
		IsDone:      s.State == "done",
		Details:     status,
		IsExecuting: strings.HasPrefix(string(s.State), "exec"),
		IsResetting: strings.HasPrefix(string(s.State), "reset"),
		Nodes:       nodes,
	}, nil
}
