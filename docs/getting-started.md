# Getting Started

This guide will walk through configuring GoAlert for general production use cases.

Most options in GoAlert are configured through the UI in the Admin page. In this guide, when configuring external services,
those options will be referenced in the format: `<Section.Option Name>` where `Section` is the section/header within the admin page and
`Option Name` refers to the label of the individual option being referenced.

The only hard requirement for GoAlert is a running Postgres instance/database.

### Running Behind a Proxy

When running GoAlert behind a reverse proxy:

- Specify the `--http-prefix` flag or `GOALERT_HTTP_PREFIX` env var for any instances behind the proxy with a path prefix _without_ the trailing slash
- Ensure the proxy passes the complete path, including prefix, if applicable
- Ensure the proxy passes the original host header (used for validating Twilio requests)
- Ensure the `General.PublicPath` contains the prefix in the URL, if applicable

## Database

We recommend using Postgres 11 for new installations as newer features will be used in the future.

GoAlert requires the `pgcrypto` extension enabled (you can enable it with `CREATE EXTENSION pgcrypto;`).
Upon first startup, it will attempt to enable the extension if it's not already enabled, but this requires elevated privileges that may not be available
in your setup.

It is also recomended to set the `--data-encryption-key` which is used to encrypt sensitive information before transmitting to the database. It can be set to any value, keep it secret.

## Running GoAlert

To run GoAlert you can start the binary or the docker container. You will need to specify the `--db-url` and `--data-encryption-key` you plan to use.

The following examples use `postgres://goalert@localhost/goalert` and `super-awesome-secret-key` respectively.

More information on Postgres connection strings can be found [here](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).

Binary:

```bash
goalert --db-url postgres://goalert@localhost/goalert --data-encryption-key super-awesome-secret-key
```

Docker:

```bash
docker run -p 8081:8081 -e GOALERT_DB_URL=postgres://goalert@localhost/goalert -e GOALERT_DATA_ENCRYPTION_KEY=super-awesome-secret-key goalert/goalert
```

You should see migrations applied followed by a `Listening.` message and an engine cycle start and end.

### API Only Mode

When running multiple instances of GoAlert (e.g. in a kubernetes cluster) it is recommended to run a single instance in the default mode, and the rest with the `--api-only` flag set.

While it is safe to run multiple "engine" instances simultaneously, it is generally unecessary and can cause unwanted contention. It is useful, however, to run an "engine" instance
in separate geographic regions or availability zones. If messages fail to send from one (e.g. network outage), they may be retried in the other this way.

## First Time Login

In order to login to GoAlert initially you will need an admin user to start with. Afterwords you may enable other authentication methods through the UI, as well as disable basic (user/pass) login.

This can be done after GoAlert has started for the first time, and is safe to execute while GoAlert is running.

To do this, you may use the `add-user` subcommand:

```bash
$ goalert add-user -h
Adds a user for basic authentication.

Usage:
  goalert add-user [flags]

Flags:
      --admin            If specified, the user will be created with the admin role (ignored if user-id is provided).
      --email string     Specifies the email address of the new user (ignored if user-id is provided).
  -h, --help             help for add-user
      --pass string      Specify new users password (if blank, prompt will be given).
      --user string      Specifies the login username.
      --user-id string   If specified, the auth entry will be created for an existing user ID. Default is to create a new user.

Global Flags:
      --data-encryption-key string       Encryption key for sensitive data like signing keys. Used for encrypting new and decrypting existing data.
      --data-encryption-key-old string   Fallback key. Used for decrypting existing data only.
      --db-url string                    Connection string for Postgres.
      --db-url-next string               Connection string for the *next* Postgres server (enables DB switch-over mode).
      --json                             Log in JSON format.
      --stack-traces                     Enables stack traces with all error logs.
  -v, --verbose                          Enable verbose logging.
```

Be sure to specify the `--admin` flag, as well as `--db-url` you plan to use.

Example usage:

```bash
goalert add-user --db-url $GOALERT_DB_URL --admin --user admin --email admin@example.com
# Prompt will be given for password
```

## Configuration

Upon logging in to GoAlert as an admin, you should see a link to the **Admin** page on the left nav-bar.

| Section | Value         | Description                                                                                                                                                                                                                                                                                     |
| ------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| General | Public URL    | Set this to the full publicly-routable URL to GoAlert. This is used for callbacks and messages from various features (like voice calls).                                                                                                                                                        |
| Auth    | Referer URLs  | By default GoAlert will only allow authentication from referers on the same host as the current request. You may manually set/restrict which hosts are allowed in environments where the UI is served by a different domain than the API or you wish to further restrict allowed referer hosts. |
| Auth    | Disable Basic | This will disable basic authentication. Do not set this until you've validated you can login as an admin by other means (e.g. GitHub or OIDC auth)                                                                                                                                              |

### GitHub Authentication

GoAlert supports GitHub's OAuth as an authentication method with the optional ability to limit logins to specified users, organizations or teams.

Follow [GitHub's documentation on creating an OAuth App](https://developer.github.com/apps/building-oauth-apps/creating-an-oauth-app/).

Using following as examples for required fields:

| Field                      | Example Value                                                    |
| -------------------------- | ---------------------------------------------------------------- |
| Application name           | `GoAlert`                                                        |
| Homepage URL               | `<General.Public URL>`                                           |
| Authorization callback URL | `<General.Public URL>/api/v2/identity/providers/github/callback` |

Document **Client ID** and **Client Secret** after creation and input into appropriate fields in GoAlert's Admin page.

Be sure to **Enable** GitHub authentication and **New Users** using the toggles and fill out **Allowed Users** or **Allowed Orgs** appropriately to restrict access.

Note: If you are limiting logins to an org or team, users will need to manually click on "Grant" access for the required org on first login (before authorizing).

### OpenID Connect Authentication (OIDC)

GoAlert supports [OpenID Connect](https://openid.net/connect/) as an authentication method.

You should be able to use any OIDC-compliant system as an authentication provider, but we'll use [Google Identity Platform using OAuth 2.0](https://developers.google.com/identity/protocols/OpenIDConnect) as an example following the **Setting up OAuth 2.0** instructions.

When creating the **user consent screen**, use the following as examples for required fields:

| Field                           | Example Value          |
| ------------------------------- | ---------------------- |
| Application name                | `GoAlert`              |
| Authorized domains              | `<General.Public URL>` |
| Application Homepage link       | `<General.Public URL>` |
| Application Privacy Policy link | `<General.Public URL>` |

When creating the **OAuth client ID**, use the following as examples for required fields:

| Field                         | Example Value                                                  |
| ----------------------------- | -------------------------------------------------------------- |
| Application type              | `Web application`                                              |
| Name                          | `GoAlert`                                                      |
| Authorized JavaScript origins | `<General.Public URL>`                                         |
| Authorized redirect URIs      | `<General.Public URL>/api/v2/identity/providers/oidc/callback` |

Document **Client ID** and **Client Secret** after creation and input into appropriate fields in GoAlert's Admin page under the **OIDC** section.

Be sure to **Enable** OIDC authentication and **New Users** using the toggles.

- Set `Override Name` to `Google` (not required).
- Set `Issuer URL` to `https://accounts.google.com`

### Mailgun

GoAlert supports creating alerts by email via Mailgun integration.

From the Admin page in GoAlert, under the `Mailgun` section, set your **Email Domain** and **API Key**.
The **API Key** may be found under the **Security** section in the Mailgun website (click your name in the top bar and select it from the drop down) it is labeled as **Private API Key**.

To configure Mailgun to forward to GoAlert:

1. Go to **Receiving**
1. Click **Create Route**
1. Set **Expression Type** to `Match Recipient`
1. Set **Recipient** to `.*@<Mailgun.Email Domain>`
1. Check **Forward**
1. In the forward box, enter `<General.Public URL>/api/v2/mailgun/incoming`
1. Click **Create Route**

### Slack

GoAlert supports generating a notification to a Slack channel or user as part of the [Escalation Policy](link to EP doc here).

For the time being you will need to create your own Slack app in your workspace for GoAlert to interface with.

To configure Slack, first [create a workspace](https://slack.com/create#email) or login to an existing one.

1. From https://api.slack.com/apps click **Create New App**
1. Enter a name for your app (e.g. `GoAlert`)
1. Select your workspace
1. Click **Create App**
1. Under **Bot Users** click **Add a Bot User** configure as desired and click **Add Bot User**
1. Under **OAuth & Permissions** add your `<General.Public URL>` with **Add New Redirect URL**
1. Under **OAuth & Permissions** click **Install App to Workspace** and complete the flow

You may now configure the **Slack** section of the Admin page.

- You may find your **Access Token** under **OAuth & Permissions** -- it is the **Bot User OAuth Access Token**
- **Client ID** and **Client Secret** are found under **Basic Information** in the **App Credentials** section.

Be sure to **Enable** Slack using the toggle.

### Twilio

GoAlert relies on bidirectional communication (outbound & inbound) with certain third-party services in order to provide convenient alerting capabilities.

For voice and SMS notifications to function, you will need a notification provider configured. Currently the only supported provider is Twilio.

Get started with a [free trial account](https://www.twilio.com/try-twilio) in order to configure GoAlert.
[Twilio's Free Trial Guide](https://support.twilio.com/hc/en-us/articles/223136107-How-does-Twilio-s-Free-Trial-work-) details the account setup instructions and limitations.
After your trial account is created, click on **Get a Trial Number** from the Twilio Dashboard.
Configure GoAlert to use your trial account by copying & pasting the following fields into the respective GoAlert fields (on the **Admin** page):

In the **Twilio** section of the Admin page:

| Twilio Dashboard | GoAlert Admin Page |
| ---------------- | ------------------ |
| TRIAL NUMBER     | From Number        |
| ACCOUNT SID      | Account SID        |
| AUTH TOKEN       | Auth Token         |

Be sure to **Enable** Twilio using the toggle.

In order for incoming SMS messages to be processed, the message callback URL must be set within Twilio.

From Twilio Dashboard, navigate to **Phone Numbers** and click on your trial phone number.

- Under **Messaging** section, update the webhook URL for _A MESSAGE COMES IN_ to `<General.Public URL>/api/v2/twilio/message`

Twilio trial account limitations (if you decide to upgrade your Twilio account these go away):

- SMS: The message "Sent from your Twilio trail account" is prepended to all SMS messages
- Voice: "You have a trial account..." verbal message before GoAlert message.
