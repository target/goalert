package swogrp

type State string

const (
	stateNeedsReset State = "needs-reset"
	stateIdle       State = "idle"
	stateReset      State = "reset"
	stateError      State = "error"
	stateExec       State = "exec"
	stateDone       State = "done"
)
