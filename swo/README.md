# Switchover (SWO)

Switchover (SWO) is a feature that allows a live system to switch from one database to another safely and with little to no user impact.

## Steps to Perform Switchover

1. **Preparation:** Configure all GoAlert instances with the `--db-url-next` flag (or using `GOALERT_DB_URL_NEXT` environment variable). Ensure that every instance is started or restarted with this configuration.

   ```
   goalert --db-url=<old-db-url> --db-url-next=<new-db-url>
   ```

   OR using environment variables:

   ```
   export GOALERT_DB_URL=<old-db-url>
   export GOALERT_DB_URL_NEXT=<new-db-url>
   ```

2. **Check Configuration in UI:**

   - Navigate to the `Admin` section in the left sidebar of the UI.
   - Click on the `Switchover` page.

3. **Initialize Switchover:**

   - While still on the `Switchover` page, click the `RESET` button. This should initiate configuration and other checks.
   - Make sure everything looks good and is validated.
   - Ensure that all instances of GoAlert (displayed as Nodes) have a green checkmark next to `Config Valid?`.

4. **Execute Switchover:**

   - Click the `EXECUTE` button to perform the database switch. It may take some time depending on DB activity and size.
   - If the operation fails (it will fail safely), click `RESET` and then `EXECUTE` again.

5. **Post-Switchover Configuration:**

   - Once the switchover is successful, reconfigure all instances to use only `--db-url`, now pointing to the **NEW** database URL.
   - Remove the `--db-url-next` flag or unset `GOALERT_DB_URL_NEXT`.

   ```
   goalert --db-url=<new-db-url>
   ```

   OR using environment variables:

   ```
   export GOALERT_DB_URL=<new-db-url>
   unset GOALERT_DB_URL_NEXT
   ```

## Development

To start the dev instance in switchover mode, run `make start-swo`

## Theory of Operation

Switchover mode is initiated by starting GoAlert with an additional DB URL `--db-url-next`. The database referenced by `--db-url` is referred to as the "old" DB and the `--db-url-next` is the "new" DB.

All new application DB connections first acquire a shared advisory lock, then check the `use_next_db` pointer. If the pointer is set, all new connections will be made to the "new" DB (without the checking overhead), and the connection to the "old" DB will be terminated.

The switch is performed by first replicating a complete snapshot of the "old" DB to the "new" DB. After the initial sync, subsequent synchronization is an incremental "diff" of snapshots -- more info on how this works is available in the `swosync` package.

After repeated logical sync operations (to keep the next-sync time low), a stop-the-world lock (i.e., an exclusive lock that conflicts with the shared advisory locks) is acquired, followed by the final logical sync. During the same transaction, the `use_next_db` pointer is set. After the lock is released, the connector will send all new queries to the "new" DB.
