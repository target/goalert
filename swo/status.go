package swo

import (
	"github.com/google/uuid"
	"github.com/target/goalert/swo/swogrp"
)

// Status represents the current status of the switchover process.
type Status struct {
	swogrp.Status

	MainDBID      uuid.UUID
	NextDBID      uuid.UUID
	MainDBVersion string
	NextDBVersion string
}
