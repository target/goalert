package graphqlapp

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/swo"
	"github.com/target/goalert/swo/swogrp"
	"github.com/target/goalert/swo/swoinfo"
)

func Test_validateSWOGrpNode(t *testing.T) {
	var s swo.Status
	s.MainDBID = uuid.New()
	s.NextDBID = uuid.New()

	// no error for valid node
	err := validateSWOGrpNode(s, swogrp.Node{OldID: s.MainDBID, NewID: s.NextDBID})
	assert.NoError(t, err)

	// should return error for invalid OldID
	err = validateSWOGrpNode(s, swogrp.Node{OldID: s.NextDBID, NewID: s.NextDBID})
	assert.Error(t, err)

	// should return error for invalid NewID
	err = validateSWOGrpNode(s, swogrp.Node{OldID: s.MainDBID, NewID: s.MainDBID})
	assert.Error(t, err)

	// should return error for invalid OldID and NewID
	err = validateSWOGrpNode(s, swogrp.Node{OldID: s.NextDBID, NewID: s.MainDBID})
	assert.Error(t, err)

	// should return error for empty OldID
	err = validateSWOGrpNode(s, swogrp.Node{OldID: uuid.Nil, NewID: s.NextDBID})
	assert.Error(t, err)

	// should return error for empty NewID
	err = validateSWOGrpNode(s, swogrp.Node{OldID: s.MainDBID, NewID: uuid.Nil})
	assert.Error(t, err)

	// should return error for empty OldID and NewID
	err = validateSWOGrpNode(s, swogrp.Node{OldID: uuid.Nil, NewID: uuid.Nil})
	assert.Error(t, err)
}

func Test_gqlSWOConnFromConnName(t *testing.T) {
	// should return unknown-conn for unknown connection
	nodeID, conn := gqlSWOConnFromConnName(swoinfo.ConnCount{Name: "foobar", Count: 1, IsNext: true})
	assert.Equal(t, "unknown-foobar", nodeID)
	assert.Equal(t, graphql2.SWOConnection{Name: "foobar", Count: 1, IsNext: true}, conn)

	nodeID, conn = gqlSWOConnFromConnName(swoinfo.ConnCount{Name: "GoAlert v0.31.0 SWO:B:AAAAAAAAAAAAAAAAAAAAAA", Count: 1, IsNext: true})
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", nodeID) // should return nodeID for valid connection
	assert.Equal(t, graphql2.SWOConnection{
		Name:    "GoAlert v0.31.0 SWO:B:AAAAAAAAAAAAAAAAAAAAAA",
		Version: "v0.31.0",
		Count:   1,
		Type:    "B",
		IsNext:  true,
	}, conn)
}

func Test_validateNodeConnections(t *testing.T) {
	err := validateNodeConnections(graphql2.SWONode{})
	assert.Error(t, err) // no connections

	err = validateNodeConnections(graphql2.SWONode{Connections: []graphql2.SWOConnection{{Name: "foobar"}}})
	assert.Error(t, err) // invalid connection

	err = validateNodeConnections(graphql2.SWONode{Connections: []graphql2.SWOConnection{{Type: string(swo.ConnTypeMainApp), IsNext: true}}})
	assert.Error(t, err) // invalid connection (wrong DB)

	err = validateNodeConnections(graphql2.SWONode{Connections: []graphql2.SWOConnection{{Type: string(swo.ConnTypeNextApp), IsNext: true}}})
	assert.NoError(t, err) // valid connection

	// test mismatched connection versions
	err = validateNodeConnections(graphql2.SWONode{Connections: []graphql2.SWOConnection{
		{Type: string(swo.ConnTypeNextApp), IsNext: true, Version: "v1.0.0"},
		{Type: string(swo.ConnTypeMainApp), IsNext: false, Version: "v1.0.1"},
	}})
	assert.Error(t, err)

	// matching versions
	err = validateNodeConnections(graphql2.SWONode{Connections: []graphql2.SWOConnection{
		{Type: string(swo.ConnTypeNextApp), IsNext: true, Version: "v1.0.0"},
		{Type: string(swo.ConnTypeMainApp), IsNext: false, Version: "v1.0.0"},
	}})
	assert.NoError(t, err)
}

func b64(id uuid.UUID) string {
	return base64.RawURLEncoding.EncodeToString(id[:])
}

func Test_gqlSWOStatus(t *testing.T) {
	node1ID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	node2ID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainDBID := uuid.New()
	nextDBID := uuid.New()

	s := swo.Status{
		MainDBID: mainDBID,
		NextDBID: nextDBID,
		Status: swogrp.Status{
			State:    swogrp.ClusterStateIdle,
			LeaderID: node2ID,
			Nodes: []swogrp.Node{
				{ID: node1ID, OldID: mainDBID, NewID: nextDBID, StartedAt: time.Now().Add(-time.Minute)},
				{ID: node2ID, CanExec: true, OldID: mainDBID, NewID: nextDBID, StartedAt: time.Now().Add(-2 * time.Minute)},
			},
		},
		MainDBVersion: "v1.0.0", // not realistic, but good enough for testing
		NextDBVersion: "v2.0.0",
	}
	conns := []swoinfo.ConnCount{
		{Name: "GoAlert v0.31.0 SWO:D:" + b64(node1ID), Count: 1, IsNext: true},
		{Name: "GoAlert v0.31.0 SWO:C:" + b64(node1ID), Count: 1, IsNext: true},
		{Name: "GoAlert v0.31.0 SWO:A:" + b64(node2ID), Count: 1, IsNext: false},
		{Name: "foobar", Count: 1, IsNext: true},
	}

	gql, err := gqlSWOStatus(s, conns)
	assert.NoError(t, err)
	assert.Equal(t, &graphql2.SWOStatus{
		State:         graphql2.SWOStateIdle,
		MainDBVersion: "v1.0.0",
		NextDBVersion: "v2.0.0",
		Nodes: []graphql2.SWONode{
			{
				ID:     node1ID.String(),
				Uptime: "1m0s",
				Connections: []graphql2.SWOConnection{
					{Name: "GoAlert v0.31.0 SWO:C:" + b64(node1ID), Version: "v0.31.0", Count: 1, Type: "C", IsNext: true},
					{Name: "GoAlert v0.31.0 SWO:D:" + b64(node1ID), Version: "v0.31.0", Count: 1, Type: "D", IsNext: true},
				},
			},
			{
				ID:       node2ID.String(),
				Uptime:   "2m0s",
				CanExec:  true,
				IsLeader: true,
				Connections: []graphql2.SWOConnection{
					{Name: "GoAlert v0.31.0 SWO:A:" + b64(node2ID), Version: "v0.31.0", Count: 1, Type: "A", IsNext: false},
				},
			},
			{
				ID: "unknown-foobar",
				Connections: []graphql2.SWOConnection{
					{Name: "foobar", Count: 1, IsNext: true},
				},
			},
		},
	}, gql)
}
