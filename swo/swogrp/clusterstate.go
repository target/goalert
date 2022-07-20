package swogrp

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
