# Using Webhooks

Webhooks are POST requests to specified endpoints with a content type of `application/json`. Webhook calls must complete within 3 seconds.

Below are example payloads:

### Verification Message

Triggered upon creating a Contact Method of type Webhook.

```
{
    "AppName": "GoAlert",
    "Type": "Verification",
    "Code": "283917"
}
```

### Test Message

Triggered on the profile page by clicking "Send Test".

```
{
    "AppName": "GoAlert",
    "Type": "Test"
}
```

### Alert

Triggered for notification of a single alert.

```
{
    "AppName": "GoAlert",
    "Type": "Alert",
    "AlertID": 79685,
    "Summary": "Example Summary",
    "Details": "Example Details..."
    "Meta": {
        "example_field": "example_value",
        "example_field2": "example_value2"
    }
}
```

### Alert Bundles

Triggered for notification of multiple alerts for a given service.

- Message Bundles must be enabled by an administrator

```
{
    "AppName": "GoAlert",
    "Type": "AlertBundle",
    "ServiceID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    "ServiceName": "Example Service",
    "Count": 6
}
```

### Status Updates

Triggered for notification of a single alert status update.

- Recipient must enable Alert Status Updates from their Profile

```
{
    "AppName": "GoAlert",
    "Type": "AlertStatus",
    "AlertID": 79694,
    "LogEntry": "Closed via test integration (Generic API)"
}
```
