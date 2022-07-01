package swosync

// txInProgressLock will cause the transaction to abort if it's unable to get
// the exec lock and/or switchover state is not currently in_progress
const txInProgressLock = `
do $$
declare
begin
	set local idle_in_transaction_session_timeout = 60000;
	set local lock_timeout = 60000;
	assert (select pg_try_advisory_xact_lock(4370)), 'failed to get exec lock';
	assert (select current_state = 'in_progress' from switchover_state), 'switchover state is not in_progress';
end $$;
`

// txInProgressLock will cause the transaction to abort if it's unable to get
// the exec lock and/or switchover state is not currently in_progress
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
	assert (select pg_try_advisory_lock(4370)), 'failed to get exec lock';
	assert (select current_state != 'use_next_db' from switchover_state), 'switchover state is use_next_db';
end $$;
`
