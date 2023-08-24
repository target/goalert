# Quick Switchover Guide for GoAlert

Switchover (SWO) is a feature that allows a live system to switch from one database to another safely and with little to no user impact.

This guide provides a quick and easy-to-follow set of steps for performing a database switchover in GoAlert.

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
   - Ensure that all instances of GoAlert (displayed as Nodes) have a green checkmark next to `Config Valid?`.

3. **Initialize Switchover:**

   - While still on the `Switchover` page, click the `RESET` button. This should initiate configuration and other checks.
   - Make sure everything looks good and is validated.

4. **Execute Switchover:**

   - Click the `EXECUTE` button to perform the database switch.
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

## Rollback Procedures

In case you encounter issues during the switchover or decide to cancel the operation, you can do so safely by following these rollback steps:

1. **Cancel Switchover in UI:**

   - Navigate to the `Switchover` page under the `Admin` section in the UI.
   - Click the `RESET` button. This will cancel the ongoing switchover process and restore the original database configuration.

2. **Reconfigure Instances:**

   - Redeploy or restart your GoAlert instances with the original `--db-url` flag, while removing the `--db-url-next` flag or the corresponding environment variable `GOALERT_DB_URL_NEXT`.

   ```
   goalert --db-url=<old-db-url>
   ```

   OR using environment variables:

   ```
   export GOALERT_DB_URL=<old-db-url>
   unset GOALERT_DB_URL_NEXT
   ```

**Note:** After a successful switchover, the old database will be marked as obsolete. GoAlert instances will refuse to start if configured to use this old database. Therefore, rollback after a successful switchover is not possible without administrative intervention.
