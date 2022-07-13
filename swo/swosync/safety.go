/*
	Locks:
	- 4919: migration lock, used to ensure only a single instance is performing migrations (or any sync operations)
	- 4369: global switchover lock, in SWO mode, all instances must acquire this lock before performing any queries
			during the switchover, an exclusive lock is acquired by the executing node (stop-the-world).
*/

package swosync

// txInProgressLock will cause the transaction to abort if it's unable to get
// the exec lock and/or switchover state is not currently in_progress
const txInProgressLock = `
do $$
declare
begin
	set local idle_in_transaction_session_timeout = 60000;
	set local lock_timeout = 60000;
	assert (select pg_try_advisory_xact_lock(4919)), 'failed to get migration lock';
	assert (select current_state = 'in_progress' from switchover_state), 'switchover state is not in_progress';
end $$;
`

// txStopTheWorld will grab the global switchover lock, halting all database activity
const txStopTheWorld = `
do $$
declare
begin
	set local idle_in_transaction_session_timeout = 3000;
	set local lock_timeout = 3000;
	perform pg_advisory_xact_lock(4369);
	assert (select current_state = 'in_progress' from switchover_state), 'switchover state is not in_progress';
end $$;
`

// ConnLockQuery will result in a failed assertion if it is unable to get the exec lock
// or switchover state is use_next_db
const ConnLockQuery = `
do $$
declare
begin
	set idle_in_transaction_session_timeout = 60000;
	set lock_timeout = 60000;
	assert (select pg_try_advisory_lock(4919)), 'failed to get migration lock';
	assert (select current_state != 'use_next_db' from switchover_state), 'switchover state is use_next_db';
end $$;
`
