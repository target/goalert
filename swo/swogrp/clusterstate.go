package swogrp

// ClusterState represents the current state of the SWO cluster.
type ClusterState int

const (
	ClusterStateUnknown ClusterState = iota
	ClusterStateResetting
	ClusterStateIdle
	ClusterStateSyncing
	ClusterStatePausing
	ClusterStateExecuting
	ClusterStateDone
)
