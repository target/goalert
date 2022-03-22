package swomsg

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Header

	Ping     *Ping     `json:",omitempty"`
	Ack      *Ack      `json:",omitempty"`
	Reset    *Reset    `json:",omitempty"`
	Execute  *Execute  `json:",omitempty"`
	Error    *Error    `json:",omitempty"`
	Plan     *Plan     `json:",omitempty"`
	Progress *Progress `json:",omitempty"`
	Done     *Done     `json:",omitempty"`
	Hello    *Hello    `json:",omitempty"`
}

type Header struct {
	ID     uuid.UUID
	NodeID uuid.UUID
	TS     time.Time `json:"-"`
}

type (
	// user commands
	Ping    struct{}
	Reset   struct{}
	Execute struct{}

	Hello struct {
		IsOldDB bool
		Status  string
		CanExec bool `json:",omitempty"`
	}

	Ack struct {
		MsgID  uuid.UUID
		Status string
		Exec   bool `json:",omitempty"`
	}

	// task updates
	Progress struct {
		MsgID   uuid.UUID
		Details string
	}
	Error struct {
		MsgID   uuid.UUID
		Details string
	}
	Done struct{ MsgID uuid.UUID }

	Plan struct {
		// Must receive Ack from all nodes before this time.
		ConsensusDeadline time.Time

		// Must receive PlanStart or Error before this time, otherwise all
		// nodes will Error.
		StartAt time.Time

		// All nodes should disable idle connections after this time.
		DisableIdleAt time.Time

		// All nodes should re-enable idle connections after this time.
		Deadline time.Time
	}
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
