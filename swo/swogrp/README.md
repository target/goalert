# SWO Group

The `swogrp` package handles orchestrating the state and transitions of the SWO process. The state of the cluster can be determined by following the sequence in the message log, which is the source of truth.

## Cluster State

```mermaid
sequenceDiagram
actor admin as Admin
participant api as API Node
participant log as Message Log
participant engine as Engine Node


note over admin,engine: Cluster State: **Unknown**
admin ->> api : Click(Reset)
activate api
api ->> log : "cancel"
api ->> api : DisableTriggers()
api ->> log : "reset"
api -->> admin: OK
deactivate api

note over admin,engine: Cluster State: **Resetting**

engine ->> log: "hello"
activate engine
note over engine: Becomes Leader
api ->> log: "hello"
engine ->> engine: Wait 3s for "hello" messages
engine ->> log: "reset-end"
deactivate engine
note over admin,engine: Cluster State: **Idle**

admin ->> api: Click(Execute)
activate api
api ->> log: "execute"
api -->> admin: OK
deactivate api

note over admin,engine: Cluster State: **Syncing**

log -->> engine: "execute"
activate engine
engine ->> engine: EnableTriggers()
engine ->> engine: InitialSync()
engine ->> engine: LogicalSync() x10
engine ->> log: "pause"
deactivate engine

note over admin,engine: Cluster State: **Pausing**
engine ->> engine: Pause()
engine ->> log: "paused"
api ->> api: Pause()
api ->> log: "paused"

note over admin,engine: Cluster State: **Executing**
log -->> engine: 2/2 "paused"
activate engine
engine ->> engine: LogicalSync() x10
engine ->> engine: FinalSync()
engine ->> log: "done"
deactivate engine

note over admin,engine: Cluster State: **Done**
```
