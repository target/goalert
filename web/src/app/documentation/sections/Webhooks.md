# Using Webhooks

Webhooks are POST requests to specified endpoints with a content type of `application/json`. Below are example payloads:

### Verification Message

Triggered upon creating a Contact Method of type Webhook.

```
{
    "Type": "Verification",
    "Code": "283917"
}
```

### Test Message

Triggered on the profile page by clicking "Send Test".

```
{
    "Type": "Test"
}
```

### Alert

Triggered for notification of a single alert.

```
{
    "Type": "Alert",
    "AlertID": 79685,
    "Summary": "Example Summary",
    "Details": "Example Details..."
}
```

### Alert Bundles

Triggered for notification of multiple alerts for a given service.

- Message Bundles must be enabled by an administrator

```
{
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
    "Type": "AlertStatus",
    "AlertID": 79694,
    "LogEntry": "Closed via test integration (Generic API)"
}
```
