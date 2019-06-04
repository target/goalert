## Getting Started

## Basic Configuration

After GoAlert is started, additional configuration may be performed from within the application by navigating to the **Admin** page.

### Exposing GoAlert (TODO) - SMS outbound doesn't require exposing -- is it worth calling this out (test SMS minus status callbacks, SMS responses, etc)?

GoAlert relies on bidirectional communication (outbound & inbound) with certain third-party services in order to provide convenient alerting capabilities.

TODO: small blurb about exposing using ngrok / serveo / reverse tunnel of your choice.

Requirements:

- Your application must be exposed externally via HTTPS (publicly routable URL)
- The steps to do this depend on multiple factors and are outside the scope of this doc -- recommend serveo / easy mode for evaluation purposes.

**PUBLIC-URL**: document and use below

### Twilio

GoAlert uses [Twilio](https://www.twilio.com/) to generate SMS and voice notifications.  
Get started with a [free trial account](https://www.twilio.com/try-twilio) in order to configure GoAlert.  
[Twilio's Free Trial Guide](https://support.twilio.com/hc/en-us/articles/223136107-How-does-Twilio-s-Free-Trial-work-) details the account setup instructions and limitations.  
After your trial account is created, click on **Get a Trial Number** from the Twilio Dashboard.  
Configure GoAlert to use your trial account by copying & pasting the following fields into the respective GoAlert fields (on the **Admin** page):

| From Twilio Dashboard | GoAlert Admin Page |
| --------------------- | ------------------ |
| TRIAL NUMBER          | From Number        |
| ACCOUNT SID           | Account SID        |
| AUTH TOKEN            | Auth Token         |

Be sure to **Enable** Twilio using the toggle.

From Twilio Dashboard, navigate to **Phone Numbers** and click on your trial phone number.

- Under **Voice & Fax** section, update the webhook URL for _A CALL COMES IN_ to **PUBLIC-URL**/api/v2/twilio/call
- Under **Voice & Fax** section, update the webhook URL for _CALL STATUS CHANGES_ to **PUBLIC-URL**/api/v2/twilio/call/status
- Under **Messaging** section, update the webhook URL for _A MESSAGE COMES IN_ to **PUBLIC-URL**/api/v2/twilio/message

Twilio trial account limitations (if you decide to upgrade your Twilio account these go away):

- SMS: The message "Sent from your Twilio trail account" is prepended to all SMS messages
- Voice: "You have a trial account..." verbal message before GoAlert message.

### Authentication

#### Basic (should we include this section or omit entirely?)

GoAlert supports basic authentication. To create an admin user, you can use the GoAlert binary:
`goalert add-user --admin --email admin@example.com --user admin --pass admin123 --db-url='postgres://goalert@/goalert?sslmode=disable'`

#### GitHub (OAuth)

GoAlert supports GitHub's OAuth as an authentication method with the optional ability to limit logins to specified users, organizations or teams.

Follow [GitHub's documentation on creating an OAuth App](https://developer.github.com/apps/building-oauth-apps/creating-an-oauth-app/).

Using following as examples for required fields:
Application name = **GoAlert**
Homepage URL = **PUBLIC-URL** (from above)
Authorization callback URL = **PUBLIC-URL/api/v2/identity/providers/github/callback**

Document **Client ID** and **Client Secret** after creation and input into appropriate fields in GoAlert's Admin page.

Be sure to **Enable** GitHub authentication and **New Users** using the toggles and fill out **Allowed Users** or **Allowed Orgs** appropriately.

- Note: If you are limiting logins to an org or team, users will need to manually click on "Grant" access for the required org on first login (before authorizing).

#### OpenID Connect (OIDC)

GoAlert supports [OpenID Connect](https://openid.net/connect/) as an authentication method. You should be able to use any OIDC-compliant system as an authentication provider, but we'll use [Google Identity Platform using OAuth 2.0](https://developers.google.com/identity/protocols/OpenIDConnect) following the **Setting up OAuth 2.0** instructions:

When creating the **user consent screen**, use the following as examples for required fields:
Application name = **GoAlert**
Authorized domains = **PUBLIC-URL** (from above)
Application Homepage link = **PUBLIC-URL** (from above)
Application Privacy Policy link = **PUBLIC-URL** (from above)

When creating the **OAuth client ID**, use the following as examples for required fields:
Application type = **Web application**
Name = **GoAlert**
Authorized JavaScript origins = **PUBLIC-URL** (from above)
Authorized redirect URIs = **PUBLIC-URL/api/v2/identity/providers/oidc/callback**

Document **Client ID** and **Client Secret** after creation and input into appropriate fields in GoAlert's Admin page.

Be sure to **Enable** OIDC authentication and **New Users** using the toggles.
Override Name = **Google**
Issuer URL = **https://accounts.google.com**

### Mailgun

Mailgun

### Slack

GoAlert supports generating a notification to a Slack channel or user as part of the [Escalation Policy](link to EP doc here).

To configure Slack, first [create a workspace](https://slack.com/create#email) and a new test channel.

https://api.slack.com/apps
Create New app

App Name: GoaLert
select workspace
create app

permissions -> "Send messages as GoAlert

chat:write:bot" save changes

Redirect URL -> new redirect url: **PUBLIC-URL**

install app to workspace

auth

copy access token in admin page

navigate to basic information
client id / secret
copy to admin page

TO DO:
cleanup github auth token stuff
cleanup google auth token stuff
cleanup slack app
cleanup mailgun stuff (DNS)
