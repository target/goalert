package swogrp

import "github.com/google/uuid"

// Status represents the current status of the switchover process.
type Status struct {
	State      ClusterState
	Nodes      []Node
	LeaderID   uuid.UUID
	LastStatus string
	LastError  string
}

// Status returns the current status of the switchover process.
func (t *TaskMgr) Status() Status {
	t.mx.Lock()
	defer t.mx.Unlock()

	nodes := make([]Node, 0, len(t.nodes))
	for _, n := range t.nodes {
		nodes = append(nodes, n)
	}

	return Status{
		State:      t.state,
		Nodes:      nodes,
		LeaderID:   t.leaderID,
		LastStatus: t.lastStatus,
		LastError:  t.lastError,
	}
}
