# Using Webhooks

Webhooks are POST requests to specified endpoints with a content type of `application/json`. Below are example payloads:

### Verification Message

```
{
    "AlertType": "Verification",
    "Summary": "Verification Message",
    "Details": "This is a verification message from GoAlert",
    "Code": "283917"
}
```

### Test Message (from profile page)

```
{
   "AlertType": "Test",
   "Summary": "Test Message",
   "Details": "This is a test message from GoAlert"
}
```

### Alert

```
{
    "AlertType": "Alert",
    "AlertID": 79685,
    "Summary": "Example Summary",
    "Details": "Example Details..."
}
```

### Alert Bundles

```
{
  "AlertType": "AlertBundle",
  "ServiceID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "ServiceName": "Example Service",
  "Count": 6
}
```

### Status Updates

```
{
    "AlertType": "AlertStatus",
    "AlertID": 79694,
    "LogEntry": "Closed via test integration (Generic API)"
}
```

### Bundled Status Updates

```
{
  "AlertType": "AlertStatusBundle",
  "AlertID": 79696,
  "Count": 2,
  "LogEntry": "Closed via test integration (Generic API)"
}
```
