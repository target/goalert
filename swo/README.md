# Switchover (SWO)

Switchover (SWO) is a feature that allows a live system to switch from one database to another safely and with little to no user impact.

## Development

To start the dev instance in switchover mode, run `make start-swo`

## Theory of Operation

Switchover mode is initiated by starting GoAlert with an additional DB URL `--db-url-next`. The database referenced by `--db-url` is referred to as the "old" DB and the `--db-url-next` is the "new" DB.

All new application DB connections first acquire a shared advisory lock, then check the `use_next_db` pointer. If the pointer is set, all new connections will be made to the "new" DB (without the checking overhead), and the connection to the "old" DB will be terminated.

The switch is performed by first replicating a complete snapshot of the "old" DB to the "new" DB. After the initial sync, subsequent synchronization is an incremental "diff" of snapshots -- more info on how this works is available in the `swosync` package.

After repeated logical sync operations (to keep the next-sync time low), a stop-the-world lock (i.e., an exclusive lock that conflicts with the shared advisory locks) is acquired, followed by the final logical sync. During the same transaction, the `use_next_db` pointer is set. After the lock is released, the connector will send all new queries to the "new" DB.
