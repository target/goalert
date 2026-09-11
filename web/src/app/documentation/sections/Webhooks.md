# Using Webhooks

Webhooks are POST requests to specified endpoints with a content type of `application/json`. Webhook calls must complete within 3 seconds.

### Restricting Destinations (Administrators)

By default, any user who can add a webhook may point it at any URL reachable from the GoAlert server, including internal or private network addresses. Administrators can restrict this from the Admin page under **Webhook**:

- **Allowed URLs**: if set, only webhook URLs matching one of the listed URL prefixes are accepted. If empty, all URLs are allowed.
- **Block Private Addresses**: if enabled, requests to private, loopback, and link-local addresses (e.g., `10.0.0.0/8`, `127.0.0.1`, `169.254.169.254`) are rejected. The check is applied at connection time, so it also covers DNS names and redirects that resolve to such addresses. If requests are routed through an HTTP proxy, destination policy must be enforced at the proxy.

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
