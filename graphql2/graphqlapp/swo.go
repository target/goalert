package graphqlapp

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/swo/swogrp"
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

var swoRx = regexp.MustCompile(`^GoAlert ([^ ]+)(?: SWO:([A-D]):(.{24}))?$`)

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

	nodes := make(map[string]*graphql2.SWONode)
	for _, conn := range conns {
		m := swoRx.FindStringSubmatch(conn.Name)
		var connType, version string
		idStr := "unknown-" + conn.Name
		if len(m) == 4 {
			version = m[1]
			connType = m[2]
			id, err := base64.URLEncoding.DecodeString(m[3])
			if err == nil && len(id) == 16 {
				var u uuid.UUID
				copy(u[:], id)
				idStr = u.String()
			}
		}
		n := nodes[idStr]
		if n == nil {
			n = &graphql2.SWONode{ID: idStr}
			nodes[idStr] = n
		}
		n.Connections = append(n.Connections, graphql2.SWOConnection{
			Name:    conn.Name,
			IsNext:  conn.IsNext,
			Version: version,
			Type:    string(connType),
			Count:   conn.Count,
		})
	}

	s := q.SWO.Status()
validateNodes:
	for _, node := range s.Nodes {
		n := nodes[node.ID.String()]
		if n == nil {
			n = &graphql2.SWONode{ID: node.ID.String()}
			nodes[node.ID.String()] = n
		}
		n.IsLeader = node.ID == s.LeaderID
		n.CanExec = node.CanExec

		up := time.Since(node.StartedAt).Truncate(time.Second).String()
		n.Uptime = &up

		if node.NewID != s.NextDBID {
			n.ConfigError = "next-db-url is invalid"
			continue
		}
		if node.OldID != s.MainDBID {
			n.ConfigError = "db-url is invalid"
			continue
		}

		if len(n.Connections) == 0 {
			n.ConfigError = "node is not connected to any DB"
			continue
		}

		version := n.Connections[0].Version
		for _, conn := range n.Connections {
			if conn.Version != version {
				n.ConfigError = "node is connected with multiple versions of GoAlert"
				continue validateNodes
			}
			if !conn.IsNext && (conn.Type != "A" && conn.Type != "B") {
				n.ConfigError = fmt.Sprintf("connected to db-url (main) with invalid type %s (expected A or B)", conn.Type)
				continue validateNodes
			}
			if conn.IsNext && (conn.Type != "C" && conn.Type != "D") {
				n.ConfigError = fmt.Sprintf("connected to next-db-url (next) with invalid type %s (expected C or D)", conn.Type)
				continue validateNodes
			}
		}
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

	var nodeList []graphql2.SWONode
	for _, n := range nodes {
		nodeList = append(nodeList, *n)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		return nodeList[i].ID < nodeList[j].ID
	})

	return &graphql2.SWOStatus{
		State: state,

		LastStatus: s.LastStatus,
		LastError:  s.LastError,
		Nodes:      nodeList,

		NextDBVersion: s.NextDBVersion,
		MainDBVersion: s.MainDBVersion,
	}, nil
}
