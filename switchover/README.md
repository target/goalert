# DB Switchover

Switchover functionality intends to replicate data to a new empty DB and coordinate all GoAlert instances to switch their active DB without downtime, loss of data, and minimal latency impact.

## High-Level Flow

1. GoAlert instances start in "switchover mode", and know about both DBs
1. A control shell validates config & state
1. New DB is migrated to be structuraly identical to the old one
1. Old DB is instrumented with a changelog
1. An initial sync is performed of all data old -> new
1. Timetable is broadcast to begin switchover
1. Pause all DB queries
1. Replicate changes since last sync
1. Unpause and use new DB

If at any time a node is introduced, config changes, or a deadline is exceeded: nodes will broadcast an abort event and resume normal operation.
All operations will end by the deadline and either the old DB or the new one (with all changes included) will be used by all nodes.

## Switchover Mode

When in switchover mode, GoAlert instances will operate with a wrapped DB driver that will determine
which DB to use for each connection that enters the pool.

GoAlert starts in "Switchover Mode" when `--db-url-next` is set.

### New Connections

1. Connect to the _old_ DB
1. Acquire a shared advisory lock
1. Check `switchover_state` for `use_next_db`
1. If set, close and return connection to _new_ db
1. If **not** set, return current connection

When the final sync is performed an exclusive advisory lock is acquired in the transaction. Since this conflicts with the shared lock, it ensures
the final sync is performed without any running queries on the old DB. When the transaction ends `switchover_state` is checked and old or new
will be used for all connections, depending on the success of the final sync.

### Node/Instance States

Viewed from logs or with the `status` command from the `switchover-shell`.

| State          | Description                                                                              |
| -------------- | ---------------------------------------------------------------------------------------- |
| starting       | Node has reset or is still starting up.                                                  |
| ready          | Node is idle and ready for instructions.                                                 |
| armed          | Node has recieved switchover timetable and is waiting for confirmation from other nodes. |
| armed-waiting  | Node is waiting for the pause phase (all known nodes confirmed)                          |
| pausing        | Node is waiting for the engine to finish pausing.                                        |
| paused-waiting | Node is paused. Engine will not run, and idle connections are disabled.                  |
| complete       | Normal operation resumed, next db is in use (`use_db_next` is set).                      |
| aborted        | Something has triggered an abort. Node has resumed normal operation.                     |

## Performing a Switchover

To perform a switchover:

1. Set/configure `--db-url-next` for all GoAlert instances
1. Run `goalert switchover-shell` with `--db-url` and `--db-url-next` set

From the switchover shell:

1. Run `reset` and wait for all nodes to be ready (use `status` or `status -w`)
1. Using `status` validate that there are no problems. You should see "No Problems Found" printed at the bottom, or a list with possible remediations.
1. Enable change tracking (for logical replication) with `enable`
1. Optionally run `sync` (it will be run as part of execute)
1. Run `execute` and confirm to perform the switchover
1. Configure all GoAlert instances to use the **new** `--db-url` and un-set `--db-url-next`

The `execute` command will ask for confirmation of the proposed timetable:

```
Switch-Over Details
  Pause API Requests: no       # Pause API requests for the full duration, instead of just the final sync
  Consensus Timeout : 3s       # Deadline for all nodes to confirm they got the timetable and are ready
  Pause Starts After: 5s       # How long to wait before begining the pause
  Pause Timeout     : 10s      # Max time to wait for all nodes to pause before aborting
  Absolute Max Pause: 13s      # The maximum possible pause time of the engine (and API requests if set above)
  Avail. Sync Time  : 1s - 11s # Indicates the possible final sync time alloted with this configuration
  Max Alloted Time  : 18s      # Max time from begining to end of the switchover process

Ready?

   Cancel
 ❯ Go!
```

Shell commands also support `-h` for extra information and options. Use `CTRL+C` to cancel an operation (like `status -w` or `sync`)

To completely reset in the event of an issue:

1. Run `disable` to remove triggers
1. Run `reset-dest` to truncate all tables in `--db-url-next`
1. Run `reset` to reset all nodes

If `execute` fails (e.g. due to a deadline) it should be safe to retry.
