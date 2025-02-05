package rotationmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/util/log"
)

type advance struct {
	id          string
	newPosition int

	silent bool
}

type rotState struct {
	ShiftStart time.Time
	Position   int
	Version    int
}

// calcAdvance will calculate rotation advancement if it is required. If not, nil is returned
func calcAdvance(ctx context.Context, t time.Time, rot *rotation.Rotation, state rotState, partCount int) *advance {
	var mustUpdate bool
	origPos := state.Position

	// get next shift start time
	newStart := rot.EndTime(state.ShiftStart)
	if state.Version == 1 {
		newStart = calcVersion1EndTime(rot, state.ShiftStart)
		mustUpdate = true
	}

	if state.Position >= partCount {
		// deleted last participant
		state.Position = 0
		mustUpdate = true
	}

	if newStart.After(t) || state.Version == 1 {
		if mustUpdate {
			return &advance{
				id:          rot.ID,
				newPosition: state.Position,

				// If migrating from version 1 to 2 without changing
				// who's on-call do so silently.
				silent: state.Version == 1 && state.Position == origPos,
			}
		}
		// in the future, so nothing to do yet
		return nil
	}

	if !newStart.After(t.Add(-15 * time.Minute)) {
		log.Log(log.WithField(ctx, "RotationID", rot.ID), fmt.Errorf("rotation advanced late (%s)", t.Sub(newStart).String()))
	}

	state.ShiftStart = newStart

	c := 0
	for {
		c++
		if c > 10000 {
			panic("too many rotation advances")
		}

		state.Position = (state.Position + 1) % partCount
		end := rot.EndTime(state.ShiftStart)
		if end.After(t) {
			break
		}
		state.ShiftStart = end
	}

	return &advance{
		id:          rot.ID,
		newPosition: state.Position,
	}
}
