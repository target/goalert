package swomsg

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Header

	Ping    *Ping    `json:",omitempty"`
	Pong    *Pong    `json:",omitempty"`
	Reset   *Reset   `json:",omitempty"`
	Error   *Error   `json:",omitempty"`
	Claim   *Claim   `json:",omitempty"`
	Execute *Execute `json:",omitempty"`
}

type Header struct {
	ID     uuid.UUID
	NodeID uuid.UUID
	TS     time.Time
}

type (
	Ping struct{}
	Pong struct{ IsNextDB bool }

	Reset   struct{ ClaimDeadline time.Time }
	Execute struct{ ClaimDeadline time.Time }

	Claim struct {
		MsgID uuid.UUID
	}

	Error struct{ Details string }

	Plan struct {
		BeginAt           time.Time
		ConsensusDeadline time.Time
		GlobalPauseAt     time.Time
		AbsoluteDeadline  time.Time
	}
	ConfirmPlan struct{ MsgID uuid.UUID }
	Progress    struct {
		Details string
	}

	Done struct{}
)

/*
UI

{ Connections Section }

{ Nodes section, with "Refresh" button}
Node ID | Ping Response Time | DB Calls/min (1m, 5m, 15m) | DB Resp. Avg (1m, 5m, 15m)

States: Idle, Error, Active, Done
{ Status section (progress text here), with "Reset", "Execute" buttons}

1. User goes to UI page

2. User clicks "Refresh" button
3. Ping is sent
4. Pong is received from all nodes
5. UI updates

6. User clicks "Execute" button
7. Execute is sent

8. Execute is claimed by engine
9. Begins instrumenting, syncing, etc... sending Progress messages
10. UI updates with progress

11. Engine sends out Plan message
12. All nodes ConfirmPlan by ConsensusDeadline
13. Engine performs switchover

14. Engine sends Done message

** if anything goes wrong, engine sends Error message and Reset is required by the user

*/
