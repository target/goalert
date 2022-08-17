package swogrp

import (
	"time"

	"github.com/google/uuid"
)

type Node struct {
	ID uuid.UUID

	CanExec bool

	OldID uuid.UUID
	NewID uuid.UUID

	StartedAt time.Time
}
