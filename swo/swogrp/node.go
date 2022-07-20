package swogrp

import "github.com/google/uuid"

type Node struct {
	ID uuid.UUID

	CanExec bool

	OldID uuid.UUID
	NewID uuid.UUID
}
