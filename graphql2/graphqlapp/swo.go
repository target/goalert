package graphqlapp

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/swo"
	"github.com/target/goalert/swo/swogrp"
	"github.com/target/goalert/swo/swoinfo"
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

// validateSWOGrpNode validates that the node has the correct DB urls.
func validateSWOGrpNode(s swo.Status, node swogrp.Node) error {
	if node.NewID != s.NextDBID {
		return fmt.Errorf("next-db-url is invalid")
	}
	if node.OldID != s.MainDBID {
		return fmt.Errorf("db-url is invalid")
	}

	return nil
}

// gwlSWOConnFromConnName maps a DB connection count to a GraphQL type.
func gqlSWOConnFromConnName(countInfo swoinfo.ConnCount) (nodeID string, conn graphql2.SWOConnection) {
	var connType, version string
	idStr := "unknown-" + countInfo.Name
	info, _ := swo.ParseConnInfo(countInfo.Name)
	if info != nil {
		version = info.Version
		connType = string(info.Type)
		idStr = info.ID.String()
	}

	return idStr, graphql2.SWOConnection{
		Name:    countInfo.Name,
		IsNext:  countInfo.IsNext,
		Version: version,
		Type:    string(connType),
		Count:   countInfo.Count,
	}
}

// validateNodeConnections ensures that the node has the correct number of connections, identified as the correct & expected type(s).
func validateNodeConnections(n graphql2.SWONode) error {
	if len(n.Connections) == 0 {
		return fmt.Errorf("node is not connected to any DB")
	}

	version := n.Connections[0].Version
	for _, conn := range n.Connections {
		if conn.Version != version {
			return fmt.Errorf("node has multiple versions: %s and %s", version, conn.Version)
		}

		if len(conn.Type) != 1 {
			return fmt.Errorf("invalid connection type: %s", conn.Type)
		}

		if conn.IsNext != swo.ConnType(conn.Type[0]).IsNext() {
			return fmt.Errorf("node has invalid connection type: %s", conn.Type)
		}
	}

	return nil
}

// gqlStateFromSWOState maps a SWO state to a GraphQL type.
func gqlStateFromSWOState(st swogrp.ClusterState) (graphql2.SWOState, error) {
	switch st {
	case swogrp.ClusterStateUnknown:
		return graphql2.SWOStateUnknown, nil
	case swogrp.ClusterStateResetting:
		return graphql2.SWOStateResetting, nil
	case swogrp.ClusterStateIdle:
		return graphql2.SWOStateIdle, nil
	case swogrp.ClusterStateSyncing:
		return graphql2.SWOStateSyncing, nil
	case swogrp.ClusterStatePausing:
		return graphql2.SWOStatePausing, nil
	case swogrp.ClusterStateExecuting:
		return graphql2.SWOStateExecuting, nil
	case swogrp.ClusterStateDone:
		return graphql2.SWOStateDone, nil
	}

	return "", fmt.Errorf("invalid state: %d", st)
}

// gqlSWOStatus maps a SWO status and connection list to a GraphQL type.
func gqlSWOStatus(s swo.Status, conns []swoinfo.ConnCount) (*graphql2.SWOStatus, error) {
	nodes := make(map[string]*graphql2.SWONode)

	// sort connections by name to ensure consistent ordering
	sort.Slice(conns, func(i, j int) bool {
		return conns[i].Name < conns[j].Name
	})
	// map connections to nodes
	for _, conn := range conns {
		idStr, c := gqlSWOConnFromConnName(conn)

		n := nodes[idStr]
		if n == nil {
			n = &graphql2.SWONode{ID: idStr}
			nodes[idStr] = n
		}
		n.Connections = append(n.Connections, c)
	}

	// update nodes from switchover_log and validate
	for _, node := range s.Nodes {
		n := nodes[node.ID.String()]
		if n == nil {
			n = &graphql2.SWONode{ID: node.ID.String()}
			nodes[node.ID.String()] = n
		}
		n.IsLeader = node.ID == s.LeaderID
		n.CanExec = node.CanExec
		n.Uptime = time.Since(node.StartedAt).Truncate(time.Second).String()

		err := validateSWOGrpNode(s, node)
		if err != nil {
			n.ConfigError = err.Error()
			continue
		}

		err = validateNodeConnections(*n)
		if err != nil {
			n.ConfigError = err.Error()
			continue
		}
	}

	// convert to list, sort by ID (for consistency)
	var nodeList []graphql2.SWONode
	for _, n := range nodes {
		nodeList = append(nodeList, *n)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})

	// map state to GraphQL type
	state, err := gqlStateFromSWOState(s.State)
	if err != nil {
		return nil, err
	}

	return &graphql2.SWOStatus{
		State: state,

		LastStatus: s.LastStatus,
		LastError:  s.LastError,
		Nodes:      nodeList,

		NextDBVersion: s.NextDBVersion,
		MainDBVersion: s.MainDBVersion,
	}, nil
}

func (q *Query) SwoStatus(ctx context.Context) (*graphql2.SWOStatus, error) {
	if q.SWO == nil {
		return nil, validation.NewGenericError("not in SWO mode")
	}

	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	conns, err := q.SWO.ConnInfo(ctx)
	if err != nil {
		return nil, err
	}

	return gqlSWOStatus(q.SWO.Status(), conns)
}
