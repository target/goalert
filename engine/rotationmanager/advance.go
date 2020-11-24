package rotationmanager

import (
	"context"
	"time"

	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/util/log"
)

type advance struct {
	id string
	t  time.Time
	p  int
}

// calcAdvance will calculate rotation advancement if it is required. If not, nil is returned
func calcAdvance(ctx context.Context, t time.Time, rot *rotation.Rotation, state rotation.State, partCount int) *advance {

	// get next shift start time
	newStart := rot.EndTime(state.ShiftStart)
	var mustUpdate bool
	if state.Position >= partCount {
		// deleted last participant
		state.Position = 0
		mustUpdate = true
	}

	if newStart.After(t) {
		if mustUpdate {
			return &advance{
				id: rot.ID,
				t:  state.ShiftStart,
				p:  state.Position,
			}
		}
		// in the future, so nothing to do yet
		return nil
	}

	warningThreshold := time.Minute * -15
	if !newStart.After(t.Add(warningThreshold)) {
		log.Debugf(ctx, "unexpected rotation advancement")
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
		id: rot.ID,
		t:  state.ShiftStart,
		p:  state.Position,
	}
}
