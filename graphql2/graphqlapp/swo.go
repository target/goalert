package graphqlapp

import (
	"context"
	"fmt"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/swo/swogrp"
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
	case graphql2.SWOActionReset:
		err = m.SWO.Reset(ctx)
	case graphql2.SWOActionExecute:
		err = m.SWO.StartExecute(ctx)
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
	for _, n := range s.Nodes {
		nodes = append(nodes, graphql2.SWONode{
			ID:       n.ID.String(),
			OldValid: n.OldID == s.MainDBID,
			NewValid: n.NewID == s.NextDBID,
			CanExec:  n.CanExec,
			IsLeader: n.ID == s.LeaderID,
		})
	}

	var state graphql2.SWOState
	switch s.State {
	case swogrp.ClusterStateUnknown:
		state = graphql2.SWOStateUnknown
	case swogrp.ClusterStateResetting:
		state = graphql2.SWOStateResetting
	case swogrp.ClusterStateIdle:
		state = graphql2.SWOStateIdle
	case swogrp.ClusterStateSyncing:
		state = graphql2.SWOStateSyncing
	case swogrp.ClusterStatePausing:
		state = graphql2.SWOStatePausing
	case swogrp.ClusterStateExecuting:
		state = graphql2.SWOStateExecuting
	case swogrp.ClusterStateDone:
		state = graphql2.SWOStateDone
	default:
		return nil, fmt.Errorf("unknown state: %d", s.State)
	}

	return &graphql2.SWOStatus{
		State: state,

		LastStatus: s.LastStatus,
		LastError:  s.LastError,

		Connections: conns,

		NextDBVersion: s.NextDBVersion,
		MainDBVersion: s.MainDBVersion,
	}, nil
}
