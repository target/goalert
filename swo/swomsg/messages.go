package swomsg

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

/*

	Idle
	Reset -> Elect(Reset)
	Elect(Reset) -> ResetRun,ResetWait -> Idle
	Execute -> Elect(Execute)
	Elect(Execute) -> ExecuteRun,ExecuteWait ->



*/

type Message struct {
	ID   uuid.UUID
	Node uuid.UUID
	TS   time.Time `json:"-"`

	Type  string
	Ack   bool            `json:",omitempty"`
	AckID uuid.UUID       `json:",omitempty"`
	Data  json.RawMessage `json:",omitempty"`
}

// type Plan struct {
// 	// Must receive Ack from all nodes before this time.
// 	ConsensusDeadline time.Time

// 	// Must receive PlanStart or Error before this time, otherwise all
// 	// nodes will Error.
// 	StartAt time.Time

// 	// All nodes should disable idle connections after this time.
// 	DisableIdleAt time.Time

// 	// All nodes should re-enable idle connections after this time.
// 	Deadline time.Time
// }

// type Type int

// const (
// 	Unknown Type = iota
// 	Execute
// 	Ping
// )

// type msg struct {
// 	Header

// 	Ping     *Ping     `json:",omitempty"`
// 	Reset    *Reset    `json:",omitempty"`
// 	Execute  *Execute  `json:",omitempty"`
// 	WaitTx   *WaitTx   `json:",omitempty"`
// 	Ack      *Ack      `json:",omitempty"`
// 	Error    *Error    `json:",omitempty"`
// 	Plan     *Plan     `json:",omitempty"`
// 	Progress *Progress `json:",omitempty"`
// 	Done     *Done     `json:",omitempty"`
// 	Hello    *Hello    `json:",omitempty"`
// }

// type Header struct {
// 	ID     uuid.UUID
// 	NodeID uuid.UUID
// 	TS     time.Time `json:"-"`
// }

// type (
// 	Start struct {
// 		Header `json:"-"`

// 		TaskName string
// 		TaskID   uuid.UUID
// 		NodeID   uuid.UUID `json:",omitempty"`
// 	}

// 	// user commands
// 	Ping struct {
// 		Header `json:"-"`
// 	}

// 	Hello struct {
// 		Header  `json:"-"`
// 		IsOldDB bool `json:",omitempty"`
// 		IsNewDB bool `json:",omitempty"`
// 		Status  string
// 		CanExec bool `json:",omitempty"`
// 	}

// 	Ack struct {
// 		Header `json:"-"`
// 		MsgID  uuid.UUID
// 	}

// 	TaskStatus struct {
// 		Header  `json:"-"`
// 		TaskID  uuid.UUID
// 		Details string
// 	}
// 	TaskDone struct {
// 		Header `json:"-"`
// 		TaskID uuid.UUID
// 		Error  string `json:",omitempty"`
// 	}

// 	Plan struct {
// 		// Must receive Ack from all nodes before this time.
// 		ConsensusDeadline time.Time

// 		// Must receive PlanStart or Error before this time, otherwise all
// 		// nodes will Error.
// 		StartAt time.Time

// 		// All nodes should disable idle connections after this time.
// 		DisableIdleAt time.Time

// 		// All nodes should re-enable idle connections after this time.
// 		Deadline time.Time
// 	}
// )

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
