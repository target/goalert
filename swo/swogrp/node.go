package swogrp

import (
	"time"

	"github.com/google/uuid"
)

// Node represents a single node in the switchover group.
type Node struct {
	ID uuid.UUID

	CanExec bool

	OldID uuid.UUID
	NewID uuid.UUID

	StartedAt time.Time
}
