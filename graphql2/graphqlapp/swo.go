package graphqlapp

import (
	"context"
	"strings"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
)

func (m *Mutation) SwoAction(ctx context.Context, action graphql2.SWOAction) (bool, error) {
	if m.SWO == nil {
		return false, validation.NewGenericError("not in SWO mode")
	}

	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return false, err
	}

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

	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	var conns []graphql2.SWOConnection
	err = sqlutil.FromContext(ctx).
		Table("pg_stat_activity").
		Select("application_name as name, count(*)").
		Where("datname = current_database()").
		Group("name").
		Find(&conns).Error
	if err != nil {
		return nil, err
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

	var errs []string
	for _, t := range s.Errors {
		errs = append(errs, t.Name+": "+t.Error)
	}

	return &graphql2.SWOStatus{
		IsIdle:      s.State == "idle",
		IsDone:      s.State == "done",
		Details:     status,
		IsExecuting: strings.HasPrefix(string(s.State), "exec"),
		IsResetting: strings.HasPrefix(string(s.State), "reset"),
		Nodes:       nodes,
		Errors:      errs,
		Connections: conns,
	}, nil
}
