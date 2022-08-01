/*
Package swosync handles the logical replication from the source DB to the destination database during switchover.

Theory of operation

All changes (INSERT, UPDATE, DELETE) are recorded by triggers in the change_log table as
table/row_id pairs, only tracking a set of changed rows (but not their point-in-time data).
The changes are then read in and applied in batches, by reading the CURRENT state of the row
from the source database and writing it to the destination database, at the time of sync.

This avoids the need to attempt to find a sequential solution to concurrent updates, as well as
intermediate row states, by only syncing the final result. It also avoids the need to record
intermediate updates.

As an example, if a row is inserted and then updated multiple times, the next sync will result in
a single insert.

The process depends on having a valid & consistent view of the source database which can be
obtained by by a serializable transaction. Since only the final state of data is used, dependency
solving/ordering for concurrent updates is not necessary.

Basic strategy:
	1. Read all changes as table and row ids
	2. Fetch row data for each changed row
	3. Insert rows from old DB that are missing in new DB, in fkey-dependency order
	4. Update rows from old DB that exist in both, in fkey-dependency order
	5. Delete rows missing from old DB that exist in new DB, in reverse-fkey-dependency order
	6. Delete synced entries from change_log table

Further Notes

It is important to keep the sync loop as tight as is possible, particularly in "final sync" mode.
When performing the final sync, the database will be locked for the full duration so no additional
changes can be made. This is necessary to ensure that the database is in a consistent state with no
leftover changes before switchover state is updated to `use_next_db`.

A commit to the source DB ensures the Serializable state of the transaction is maintained, and is
done AFTER sending changes to the new DB as the final sync also points to the new one.

Round Trips:
	- 1 to start tx, read all change ids & sequences (also stop-the-world lock in final mode)
	- 1 to fetch row data from each table (single batch, 1 query per table)
	- 1 to apply all updates to new DB
	- 1 to commit src tx (also switches over to new DB in final mode)
	- 1 to delete all synced change rows from the DB

There is an extra round-trip for last delete as a tradoff to favor shorter stop-the-world time,
since deleting the last set of changes isn't necessary to wait for after the switchover has been made.
*/
package swosync
