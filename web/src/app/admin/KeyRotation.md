1. Restart/redeploy _ALL_ GoAlert instances with:

   - `GOALERT_DATA_ENCRYPTION_KEY` set to the **new** value.
   - `GOALERT_DATA_ENCRYPTION_KEY_OLD` set to the **old** value.

2. Once the deployment is complete, run the `RE-ENCRYPT DATA` command below.
3. Restart/redeploy _ALL_ GoAlert instances with only `GOALERT_DATA_ENCRYPTION_KEY` set to the **new** value.
