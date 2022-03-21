# Switchover (SWO)

Switchover (SWO) is a feature that allows a live system to safely switch from one database to another.

## Theory of Operation

During SWO, 2 DB url's are involved. "old" and "new".

- All app-related DB connections acquire a shared advisory lock `GlobalSwitchOver` to the "old" DB, followed by checking switchover state is not `use_next_db`. These locks are at the session level and persist as long as the connections remain in the pool.
- If it is `use_next_db`, SWO is complete, the connection is closed, and future connections are made to the "new" DB without the lock.
- Once initiated, the first engine instance to acquire the `GlobalSwitchOverExec` lock (separate from `GlobalSwitchOver`) will begin the switch.
- When the switch is started, a `change_log` table is created and populated by triggers added to existing tables for INSERT/UPDATE/DELETE operations.
- An initial sync is performed effectively copying a snapshot of all data from the "old" DB to the "new" DB.
- Subsequent syncs are performed by applying records from the `change_log` table to the "new" DB.
- After each sync, the synced rows are deleted from the `change_log` table, so that it always represents the diff between both DBs.
- This is repeated until the `change_log` table has less than 100 rows at the start of a sync.
- Once the DBs are relatively similar, SWO goes into "critical phase".
- In this phase, idle connections are disabled until a shared deadline (meaning each query requires the shared lock, as connections are not re-used).
- When the final sync begins, an exclusive `GlobalSwitchOver` lock is acquired, and behaves as a stop-the-world lock.
- After the final sync, sequences are also copied from the "old" DB to the "new" DB.
- Finally, the `current_state` column is updated to `use_next_db`, and the `GlobalSwitchOver` lock is released.

If deadlines are reached, or any error is encountered, the connection for the switchover is dropped, and syncing resumes. If a commit to the new DB succeeds, but fails on the old DB, an error state is entered.

From an error state, only RESET can be performed, which wipes the "new" DB and recreates `change_log` and all triggers to begin again with another attempt.

