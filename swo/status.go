package swo

import (
	"github.com/google/uuid"
	"github.com/target/goalert/swo/swogrp"
)

type Status struct {
	swogrp.Status

	MainDBID      uuid.UUID
	NextDBID      uuid.UUID
	MainDBVersion string
	NextDBVersion string
}
