# Grafana Development

To test the Grafana integration, ensure you have a service with a configured integration key and click "Copy Grafana Webhook URL" to add the webhook URL to your clipboard.

## Grafana 7

1. Start Grafana: `docker run -it --rm -p 3000:3000 grafana/grafana:7.5.13`
2. Login at http://localhost:3000 with `admin/admin`
3. Add a datasource
   1. Hover over gear icon
   2. Click Data sources
   3. Click Add data source
   4. Select TestData DB
   5. Click Save & Test
4. Add a notification channel
   1. Hover over bell icon
   2. Click Notification Channels
   3. Click Add Channel
   4. name=test, type=webhook, url=paste from GoAlert
   5. Click Test and ensure an alert is created (you can close it)
   6. Click Save
5. Add a Dashboard
   1. Hover over "+" Plus icon
   2. Click Dashboard
   3. Click Add an Empty Panel
   4. Click Save
6. Add alert
   1. Click panel title
   2. Edit
   3. Scenario=Predictable Pulse, Step=1, on count=10, off count=10
   4. Click Alert tab
   5. Click Create Alert
   6. Evaluate every=1s, For=3s, WHEN=last(), IS ABOVE=1.5, Send to=test
   7. Click Apply at the top of the page
   8. Click the Save icon at the top of the page
7. Alerts should be created and closed regularly

## Grafana 8

1. Start Grafana: `docker run -it --rm -p 3000:3000 grafana/grafana:8.3.4`
2. Login at http://localhost:3000 with `admin/admin`
3. Add a datasource
   1. Hover over gear icon
   2. Click Data sources
   3. Click Add data source
   4. Select TestData DB
   5. Click Save & Test
4. Add a Contact Point
   1. Hover over bell icon
   2. Click Contact Points
   3. Click New Contact Point
   4. name=test, Contact point type=webhook, Url=paste from GoAlert
   5. Click Test, then Send Test Notification and ensure an alert is created (you can close it)
   6. Click Save contact point
5. Create a folder
   1. Hover over "+" Icon
   2. Click Folder
   3. Folder name=default
   4. Click Create
6. Add alert rule
   1. Hover over bell icon
   2. Click Alert rules
   3. Click New alert rule
   4. Rule name=test, rule type=Grafana managed alert, folder=default
   5. Scenario=predictable pulse, Step=1, on count=30, off count=30
   6. IS ABOVE=1.5
   7. Evaluate every=10s, for=15s
   8. summary=some alert summary
   9. Click Save & Exit
7. Update notification policy
   1. Hover over bell icon
   2. Click Notification policies
   3. Click Edit for root policy
   4. Default contact point=test
   5. Click Save
8. Alerts should be created and closed regularly
