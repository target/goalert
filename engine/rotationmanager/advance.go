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
}

type rotState struct {
	ShiftStart time.Time
	Position   int
	Version    int
}

// calcAdvance will calculate rotation advancement if it is required. If not, nil is returned
func calcAdvance(ctx context.Context, t time.Time, rot *rotation.Rotation, state rotState, partCount int) (*advance, error) {
	var mustUpdate bool

	if state.Position >= partCount {
		mustUpdate = true
		state.Position = 0
	}

	endTimeFunc := rot.EndTime

	switch state.Version {
	case 1:
		// for a V1 state, use the old calculation method
		mustUpdate = true
		endTimeFunc = func(t time.Time) time.Time {
			return calcVersion1EndTime(rot, t)
		}
	case 2:
		// no-op
	default:
		return nil, fmt.Errorf("unknown rotation state version (supported: 1,2): %d", state.Version)
	}

	// get next shift start time
	newStart := endTimeFunc(state.ShiftStart)
	if newStart.After(t) {
		if mustUpdate {
			// we need to update the rotation state to v2, which will reset the start time and make it compatible with v2, but we
			// don't need to change the position (unless it was due to participant deletion).
			return &advance{
				id:          rot.ID,
				newPosition: state.Position,
			}, nil
		}

		// in the future, so nothing to do yet
		return nil, nil
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
		end := endTimeFunc(state.ShiftStart)
		if end.After(t) {
			break
		}
		state.ShiftStart = end
	}

	return &advance{
		id:          rot.ID,
		newPosition: state.Position,
	}, nil
}
