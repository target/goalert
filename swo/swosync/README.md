# Logical Sync

Package `swosync` handles the logical replication from the source DB to the destination database during switchover.

## Theory of operation

Triggers record all changes (INSERT, UPDATE, DELETE) in the `change_log` table as `table, row_id` pairs, only tracking a set of changed rows but not their point-in-time data. The changes are then read in and applied in batches by reading the CURRENT state of the row from the source database and writing it to the destination database at the time of sync.

Replicating point-in-time differences between "snapshots" avoids the need for a sequential solution for concurrent updates and intermediate row states by only syncing the final result. It also becomes more efficient because each row must be replicated at most once, even when multiple updates occur between sync points.

The process depends on having a consistent view of the source database, which a serializable transaction can obtain, or during a stop-the-world lock (during the final sync).

### Basic strategy

1. Read all changes (table and row ids)
2. Fetch row data for each changed row
3. Insert rows from old DB that are missing in new DB, in fkey-dependency order
4. Update rows from old DB that exist in both, in fkey-dependency order
5. Delete rows missing from the old DB that exist in the new DB, in reverse-fkey-dependency order
6. Delete synced entries from the `change_log` table
7. Repeat until both DBs are close in sync
8. Obtain a stop-the-world lock
9. Perform final sync, and update the `use_next_db` pointer
10. Release the stop-the-world lock
11. New DB is used for all future transactions

## Further Notes

It is essential to keep the sync loop as tight as possible, particularly in "final sync" mode. The final sync will pause all transactions during its synchronization process; this is necessary to ensure that the database is in a consistent state with no leftover changes before setting the `use_next_db` pointer.

### Round Trips

- 1 to start tx, read all change ids & sequences (also stop-the-world lock in final mode)
- 1 to fetch row data from each table (single batch, 1 query per table)
- 1 to apply all updates to the new DB
- 1 to commit src tx (also updates `use_next_db` in final mode)
- 1 to delete all synced change rows from the DB

An extra round-trip for the last delete is a trade-off to favor a shorter stop-the-world time since deleting the previous change records isn't necessary after the switchover.
