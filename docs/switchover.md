# Quick Switchover Guide for GoAlert

This guide provides a quick and easy-to-follow set of steps for performing a database switchover in GoAlert. For a detailed understanding, see the [README in the swo package](../swo/README.md).

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
