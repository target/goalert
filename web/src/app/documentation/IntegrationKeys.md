# API Reference

- [Email](#Email)
- [Generic API](#Generic_API)
- [Grafana](#Grafana)

---

## Email

It is now possible to create an Email integration key from the Service Details page. This will generate a unique email address that can be used for creating alerts.

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

---

## Generic API

### Params can be in query params or body (body takes precedence):

`summary` -> required, sent as the sms and voice messages  
`details` -> optional, additional information about the alert (e.g. links and whatnot)  
`action` -> optional, if set to `close`, it will close any matching alerts  
`dedup` -> optional, all calls for the same service with the same "dedup" string will update the same alert (if open) or create a new one. Defaults to using summary & details together.  
`token` -> the integration key to use

### Examples:

```bash
curl -XPOST https://<example.goalert.me>/api/v2/generic/incoming?token=key-here&summary=test
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
