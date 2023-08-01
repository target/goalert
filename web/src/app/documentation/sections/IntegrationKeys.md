# Using Integration Keys

## Generic API

### Params can be in query params or body (body takes precedence):

| Name      |              | Description                                                                                                                                                         |
| --------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `token`   | **Required** | The integration key to use.                                                                                                                                         |
| `summary` | **Required** | Short description of the alert sent as SMS and voice.                                                                                                               |
| `details` | _optional_   | Additional information about the alert, supports markdown.                                                                                                          |
| `action`  | _optional_   | If set to `close`, it will close any matching alerts.                                                                                                               |
| `dedup`   | _optional_   | All calls for the same service with the same `dedup` string will update the same alert (if open) or create a new one. Defaults to using summary & details together. |

### Response:

Default response is a `204` with no content. Set the `Accept` header to `application/json` for information on the created alert.

Example:

```json
{
  "AlertID": 10,
  "ServiceID": "00000000-0000-0000-0000-000000000001",
  "IsNew": true
}
```

`IsNew` will be false if the call was de-duplicated.

### Examples:

```bash
curl -XPOST https://<example.goalert.me>/api/v2/generic/incoming?token=key-here&summary=test&details=test
curl -XPOST https://<example.goalert.me>/api/v2/generic/incoming?token=key-here&summary=test&dedup=disk-check
curl -XPOST https://<example.goalert.me>/api/v2/generic/incoming?token=key-here&summary=test&action=close
```

---

## Grafana

Grafana provides basic alerting functionality for metrics.

To trigger an alert using Grafana, follow these steps:

1. Within GoAlert, on the Services page, select the service you want to process the alert. Under Integration Keys:

   - Key Name: Enter a name for the key.
   - Key Type: Grafana
   - Click Add Key. Copy the generated URL and keep it handy, as you'll need it in a future step.

2. In Grafana, click the Grafana icon in the top left and select Alerting > Notification Channels, then click New Channel:

   - Name: Choose a name that makes sense to people outside of your team.
   - Type: webhook
   - Send on all alerts: Do not select this checkbox unless you want to get paged for every single alert on Visualize.
   - Url: Paste in the Grafana webhook URL you generated in step 1.
   - Http Method: POST
   - Click Send Test to verify that your configuration is correct, then click Save.

3. Navigate to any of your graph panels on a dashboard, edit the panel, and click the Alert tab. Configure your alerts (if you haven't already), then in the Notifications section of the Alert tab, find the notification channel you just created in the Send to field. Click Save.

---

## Site24x7

Site24x7 provides alerting functionality for checks as an IT automation.

To trigger an alert using Site24x7, follow these steps:

1. Within GoAlert, on the Services page, select the service you want to process the alert. Under Integration Keys:

   - Key Name: Enter a name for the key.
   - Key Type: Site24x7
   - Click Add Key. Copy the generated URL and keep it handy, as you'll need it in a future step.

2. In Site24x7, go to Admin > IT Automation, then click Add Automation:

   - Display Name: Choose a name that makes sense to people outside of your team.
   - Url: Paste in the Site24x7 webhook URL you generated in step 1.
   - HTTP Method: POST
   - Send Incident Parameters: Select this to pass the details of the alert.
   - Post as JSON: Select this to pass the details in the required JSON format.
   - Click Save, testing the IT Automation will fail as the test doesn't pass any of the required information.

3. Navigate to the configuration page for the check you want to alert on, in the IT Automation section select the Automation you created above and select when you want the action to trigger.

---

## Prometheus Alertmanager

Prometheus Alertmanager provides alerting functionality for checks as an IT automation.

To trigger an alert using Prometheus Alertmanager, follow these steps:

1. Within GoAlert, on the Services page, select the service you want to process the alert. Under Integration Keys:

   - Key Name: Enter a name for the key.
   - Key Type: Prometheus Alertmanager
   - Click Add Key. Copy the generated URL and keep it handy, as you'll need it for the next step.

2. In Prometheus Alertmanager, enable a webhook by adding a webhook receiver in the alertmanager configuration file:

   ```yaml
   receivers:
     - name: 'service'
       webhook_configs:
         - url: '<prometheus_alertmanager_webhook_url_from_previous_step>'
           send_resolved: true
   ```

---

## Email

It is possible to create an Email integration key from the Service Details page. This will generate a unique email address that can be used for creating alerts.

De-duplication happens by matching subject and body contents automatically. The email subject line will become the alert summary.

You can override de-duplication if needed and use a custom key by adding
`+some_value here`
before the "@" symbol. De-duplication behaves similarly to the Grafana and generic API integration keys: if there is an open alert, "duplicate suppressed" is logged, otherwise a new alert is created.

### Custom Deduplication example

`b3b16257-75e0-4b9f-9436-db950ec0436c@target.goalert.me`
would become
`b3b16257-75e0-4b9f-9436-db950e 0436c+some_value_here@target.goalert.me`
which would match alerts created for the same service, to the same
`some_value_here`
key, regardless of the subject or body.
On the Service page, Add an Integration Key, select Email and SAVE Copy the Email address and use this with the email-based service that you want to alert on.
