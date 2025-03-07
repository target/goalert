package rotation

import "github.com/google/uuid"

// Update is an event triggered when a rotation is updated.
type Update struct {
	ID uuid.UUID
}
