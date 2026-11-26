# Microsoft Teams

GoAlert can post alert notifications, status updates, on-call notifications, and test messages to Microsoft Teams channels as Adaptive Cards.

Messages are delivered through a Power Automate Workflow webhook (the supported replacement for the retired Office 365 connectors).

### Creating a Workflow Webhook URL

1. In Microsoft Teams, open the channel you want notifications in and choose **Workflows** (or start from the Power Automate "Post to a channel when a webhook request is received" template).
2. Create the workflow **"Post to a channel when a webhook request is received"** and select the target team and channel.
3. Copy the generated HTTP POST URL.
4. In GoAlert, create a notification destination of type **Microsoft Teams** (e.g. as an escalation policy step or schedule on-call notification) and paste the URL.

### Notes

- Cards are posted by the workflow's Flow bot identity, not by GoAlert itself.
- Alert status updates are posted as new cards; existing cards are not edited.
- Acknowledge/close from within Teams is not supported with workflow webhooks; use the card's **Open Alert** button to act on an alert in GoAlert.
- Treat the workflow webhook URL as a secret: anyone with the URL can post to the channel.
- Administrators may restrict allowed webhook domains with the `Teams.AllowedWorkflowURLs` setting.

### Payload

Each message is a standard Teams message envelope containing a single Adaptive Card attachment:

```
{
    "type": "message",
    "attachments": [
        {
            "contentType": "application/vnd.microsoft.card.adaptive",
            "contentUrl": null,
            "content": { ...Adaptive Card... }
        }
    ]
}
```
