# Using Webhooks

Webhooks are POST requests to specified endpoints with a content type of `application/json`. Below are example payloads:

### Verification Message

```
{
    "Type": "Verification",
    "Summary": "Verification Message",
    "Details": "This is a verification message from GoAlert",
    "Code": "283917"
}
```

### Test Message (from profile page)

```
{
   "Type": "Test",
   "Summary": "Test Message",
   "Details": "This is a test message from GoAlert"
}
```

### Alert

```
{
    "Type": "Alert",
    "AlertID": 79685,
    "Summary": "Example Summary",
    "Details": "Example Details..."
}
```

### Alert Bundles

```
{
  "Type": "AlertBundle",
  "ServiceID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "ServiceName": "Example Service",
  "Count": 6
}
```

### Status Updates

```
{
    "Type": "AlertStatus",
    "AlertID": 79694,
    "LogEntry": "Closed via test integration (Generic API)"
}
```

### Bundled Status Updates

```
{
  "Type": "AlertStatusBundle",
  "AlertID": 79696,
  "Count": 2,
  "LogEntry": "Closed via test integration (Generic API)"
}
```
