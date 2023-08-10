# Getting Started

This guide will walk through configuring GoAlert for general production use cases.

Most options in GoAlert are configured through the UI in the Admin page. In this guide, when configuring external services,
those options will be referenced in the format: `<Section.Option Name>` where `Section` is the section/header within the admin page and
`Option Name` refers to the label of the individual option being referenced.

The only hard requirement for GoAlert is a running Postgres instance/database.

### Running Behind a Proxy

When running GoAlert behind a reverse proxy, make sure the `--public-url` includes the prefix path, if applicable. Ensure the proxy does _not_ trim the prefix before passing the request to GoAlert; it will be handled internally.

## Database

We recommend using Postgres 13 (or newer) for new installations as newer features will be used in the future.

GoAlert requires the `pgcrypto` extension enabled (you can enable it with `CREATE EXTENSION pgcrypto;`).
Upon first startup, it will attempt to enable the extension if it's not already enabled, but this requires elevated privileges that may not be available
in your setup.

Note: If you are using default install of Postgres on Debian (maybe others) you may run into an issue where the OOM (out of memory) killer terminates the supervisor process. More information along with steps to resolve can be found [here](https://www.postgresql.org/docs/current/kernel-resources.html#LINUX-MEMORY-OVERCOMMIT).

### Encryption of Sensitive Data

It is also recommended to set the `--data-encryption-key` which is used to encrypt sensitive information (like API keys) before transmitting to the database.

It can be set to any value as it is internally passed through a key derivation function. All instances of GoAlert must be configured to use the same key for things to work properly.

## Running GoAlert

To run GoAlert, you can start the binary directly, or from a container image. You will need to specify the `--db-url`, `--public-url`, and `--data-encryption-key` you plan to use.

The following examples use `postgres://goalert@localhost/goalert` and `super-awesome-secret-key` respectively.

More information on Postgres connection strings can be found [here](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).

Binary:

```bash
goalert --db-url postgres://goalert@localhost/goalert --data-encryption-key super-awesome-secret-key --public-url https://goalert.example.com
```

Container:

```bash
podman run -p 8081:8081 -e GOALERT_DB_URL=postgres://goalert@localhost/goalert -e GOALERT_DATA_ENCRYPTION_KEY=super-awesome-secret-key -e GOALERT_PUBLIC_URL=https://goalert.example.com goalert/goalert
```

You should see migrations applied followed by a `Listening.` message and an engine cycle start and end.

### API Only Mode

When running multiple instances of GoAlert (e.g. in a kubernetes cluster) it is recommended to run a single instance in the default mode, and the rest with the `--api-only` flag set.

While it is safe to run multiple "engine" instances simultaneously, it is generally unnecessary and can cause unwanted contention. It is useful, however, to run an "engine" instance
in separate geographic regions or availability zones. If messages fail to send from one (e.g. network outage), they may be retried in the other this way.

## First Time Login

In order to log in to GoAlert initially you will need an admin user to start with. Afterwords you may enable other authentication methods through the UI, as well as disable basic (user/pass) login.

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
      --db-url-next string               Connection string for the *next* Postgres server (enables DB switchover mode).
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

Upon logging in to GoAlert as an admin, you should see a link to the **Admin** page on the left nav-bar. The primary page in this section is Config and allows configuration of various providers and options.

### GitHub Authentication

GoAlert supports GitHub's OAuth as an authentication method with the optional ability to limit logins to specified users, organizations or teams.

Follow [GitHub's documentation on creating an OAuth App](https://developer.github.com/apps/building-oauth-apps/creating-an-oauth-app/).

Using following as examples for required fields:

| Field                      | Example Value                                                    |
| -------------------------- | ---------------------------------------------------------------- |
| Application name           | `GoAlert`                                                        |
| Homepage URL               | `<GOALERT_PUBLIC_URL>`                                           |
| Authorization callback URL | `<GOALERT_PUBLIC_URL>/api/v2/identity/providers/github/callback` |

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
| Authorized domains              | `<GOALERT_PUBLIC_URL>` |
| Application Homepage link       | `<GOALERT_PUBLIC_URL>` |
| Application Privacy Policy link | `<GOALERT_PUBLIC_URL>` |

When creating the **OAuth client ID**, use the following as examples for required fields:

| Field                         | Example Value                                                  |
| ----------------------------- | -------------------------------------------------------------- |
| Application type              | `Web application`                                              |
| Name                          | `GoAlert`                                                      |
| Authorized JavaScript origins | `<GOALERT_PUBLIC_URL>`                                         |
| Authorized redirect URIs      | `<GOALERT_PUBLIC_URL>/api/v2/identity/providers/oidc/callback` |

Document **Client ID** and **Client Secret** after creation and input into appropriate fields in GoAlert's Admin page under the **OIDC** section.

Be sure to **Enable** OIDC authentication and **New Users** using the toggles.

- Set `Override Name` to `Google` (not required).
- Set `Issuer URL` to `https://accounts.google.com`

**Note:** An application like [Dex](https://dexidp.io/) can be used to integrate with many other auth systems and provide an OIDC method for GoAlert.

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
1. In the forward box, enter `<GOALERT_PUBLIC_URL>/api/v2/mailgun/incoming`
1. Click **Create Route**

### Email/SMTP

GoAlert supports creating alerts from emails received directly as a SMTP server (as an alternative to using Mailgun).

You'll need to create an MX-type DNS record for your GoAlert server.

Note: This SMTP server only handles incoming messages/generating alerts; it will not send outgoing mail. For notifying users via email, you can configure an external SMTP server integration in the admin UI.

To enable the built-in SMTP ingress, pass the `--smtp-listen` and/or `--smtp-listen-tls` flag with the address to listen on, e.g. `--smtp-listen=0.0.0.0:9025`. You may use both flags with different ports.  You must also provide a domain name with `--email-integration-domain`, to be used for generating the alert-creating email addresses.

If you use the TLS variant, you must also pass the TLS cert and key. To do so, use either `--smtp-tls-cert-file` and `--smtp-tls-key-file` (with paths to the cert and key files) or `--smtp-tls-cert-data` and `--smtp-tls-key-data` to pass the cert and key data directly as strings. The cert and key must be PEM-encoded.

By default, only the `--email-integration-domain` will be allowed for the TO address on incoming emails.  To allow other domains, you can pass a comma-separated list of domains to `--smtp-additional-domains`, e.g. `--smtp-allowed-domains="example.com,foo.io"`. Messages addressed to domains not matching the given list will be rejected.

### Slack

GoAlert supports generating a notification to a Slack channel as part of the Escalation Policy.

For the time being you will need to create your own Slack app in your workspace for GoAlert to interface with.

To configure Slack, first [create a workspace](https://slack.com/create#email) or log in to an existing one.

1. Open the **Slack** section of the GoAlert Admin page
2. Click `App Manifest`
3. Click `Create New App`
4. Follow the prompts to install the app in your workspace

You may now configure the **Slack** section of the GoAlert Admin page.

- You may find your **Access Token** under **OAuth & Permissions** -- it is the **Bot User OAuth Access Token**
- **Client ID**, **Client Secret**, and **Signing Secret** are found under **Basic Information** in the **App Credentials** section.

Be sure to **Enable** Slack using the toggle.

You must invite the new app (e.g. GoAlert) by typing `/invite @GoAlert` in the desired Slack channel(s).

To have `Interactive Messages` work, you will need to link Slack and GoAlert users using a tool like `goalert-slack-email-sync` in this repo. This will be made easier (e.g., user-initiated) in the future.

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

- Under **Messaging** section, update the webhook URL for _A MESSAGE COMES IN_ to `<GOALERT_PUBLIC_URL>/api/v2/twilio/message`

Twilio trial account limitations (if you decide to upgrade your Twilio account these go away):

- SMS: The message "Sent from your Twilio trial account" is prepended to all SMS messages
- Voice: "You have a trial account..." verbal message before GoAlert message.

### CLI Flags

Additional options are available for running GoAlert in the form of CLI flags. Their corresponding environment variable names are listed as well.

| Flag                         | Environment Variable               | Description                                                                                                                                             |
| ---------------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`--db-url`**               | `GOALERT_DB_URL`                   | Connection string for Postgres.                                                                                                                         |
| **`--db-url-next`**          | `GOALERT_DB_URL_NEXT`              | Connection string for the _next_ Postgres server (enables DB switchover mode).                                                                          |
| **`--public-url`**           | `GOALERT_PUBLIC_URL`               | Externally routable URL to the application. Used for validating callback requests, links, auth, and prefix calculation.                                 |
| `--data-encryption-key`      | `GOALERT_DATA_ENCRYPTION_KEY`      | Used to generate an encryption key for sensitive data like signing keys. Can be any length. only use this when performing a switchover.                 |
| `--data-encryption-key-old`  | `GOALERT_DATA_ENCRYPTION_KEY_OLD`  | Fallback key. Used for decrypting existing data only. only necessary when changing --data-encryption-key.                                               |
| `--api-only`                 | `GOALERT_API_ONLY`                 | Starts in API-only mode (schedules & notifications will not be processed). Useful in clusters.                                                          |
| `--db-max-idle`              | `GOALERT_DB_MAX_IDLE`              | Max idle DB connections. (default 5)                                                                                                                    |
| `--db-max-open`              | `GOALERT_DB_MAX_OPEN`              | Max open DB connections. (default 15)                                                                                                                   |
| `--disable-https-redirect`   | `GOALERT_DISABLE_HTTPS_REDIRECT`   | Disable automatic HTTPS redirects.                                                                                                                      |
| `--engine-cycle-time`        | `GOALERT_ENGINE_CYCLE_TIME`        | Time between engine cycles. (default 5s)                                                                                                                |
| `--experimental`             | `GOALERT_EXPERIMENTAL`             | Enable experimental features.                                                                                                                           |
| `--github-base-url`          | `GOALERT_GITHUB_BASE_URL`          | Base URL for GitHub auth and API calls.                                                                                                                 |
| `--help`                     | -                                  | Help about any command                                                                                                                                  |
| `--json`                     | `GOALERT_JSON`                     | Log in JSON format.                                                                                                                                     |
| `--list-experimental`        | `GOALERT_LIST_EXPERIMENTAL`        | List experimental features.                                                                                                                             |
| `--listen`                   | `GOALERT_LISTEN`                   | Listen address:port for the application. (default "localhost:8081")                                                                                     |
| `--listen-prometheus`        | `GOALERT_LISTEN_PROMETHEUS`        | Bind address for Prometheus metrics.                                                                                                                    |
| `--listen-sysapi`            | `GOALERT_LISTEN_SYSAPI`            | (Experimental) Listen address:port for the system API (gRPC).                                                                                           |
| `--listen-tls`               | `GOALERT_LISTEN_TLS`               | HTTPS listen address:port for the application. Requires setting --tls-cert-data and --tls-key-data OR --tls-cert-file and --tls-key-file.               |
| `--log-engine-cycles`        | `GOALERT_LOG_ENGINE_CYCLES`        | Log start and end of each engine cycle.                                                                                                                 |
| `--log-errors-only`          | `GOALERT_LOG_ERRORS_ONLY`          | Only log errors (superseeds other flags).                                                                                                               |
| `--log-requests`             | `GOALERT_LOG_REQUESTS`             | Log all HTTP requests. If false, requests will be logged for debug/trace contexts only.                                                                 |
| `--max-request-body-bytes`   | `GOALERT_MAX_REQUEST_BODY_BYTES`   | Max body size for all incoming requests (in bytes). Set to 0 to disable limit. (default 262144)                                                         |
| `--max-request-header-bytes` | `GOALERT_MAX_REQUEST_HEADER_BYTES` | Max header size for all incoming requests (in bytes). Set to 0 to disable limit. (default 4096)                                                         |
| `--region-name`              | `GOALERT_REGION_NAME`              | Name of region for message processing (case sensitive). Only one instance per-region-name will process outgoing messages. (default "default")           |
| `--slack-base-url`           | `GOALERT_SLACK_BASE_URL`           | Override the Slack base URL.                                                                                                                            |
| `--stack-traces`             | `GOALERT_STACK_TRACES`             | Enables stack traces with all error logs.                                                                                                               |
| `--status-addr`              | `GOALERT_STATUS_ADDR`              | Open a port to emit status updates. Connections are closed when the server shuts down. Can be used to keep containers running until GoAlert has exited. |
| `--strict-experimental`      | `GOALERT_STRICT_EXPERIMENTAL`      | Fail to start if unknown experimental features are specified.                                                                                           |
| `--stub-notifiers`           | `GOALERT_STUB_NOTIFIERS`           | If true, notification senders will be replaced with a stub notifier that always succeeds (useful for staging/sandbox environments).                     |
| `--sysapi-ca-file`           | `GOALERT_SYSAPI_CA_FILE`           | (Experimental) Specifies a path to a PEM-encoded certificate(s) to authorize connections from plugin services.                                          |
| `--sysapi-cert-file`         | `GOALERT_SYSAPI_CERT_FILE`         | (Experimental) Specifies a path to a PEM-encoded certificate to use when connecting to plugin services.                                                 |
| `--sysapi-key-file`          | `GOALERT_SYSAPI_KEY_FILE`          | (Experimental) Specifies a path to a PEM-encoded private key file use when connecting to plugin services.                                               |
| `--tls-cert-data`            | `GOALERT_TLS_CERT_DATA`            | Specifies a PEM-encoded certificate. Has no effect if --listen-tls is unset.                                                                            |
| `--tls-cert-file`            | `GOALERT_TLS_CERT_FILE`            | Specifies a path to a PEM-encoded certificate. Has no effect if --listen-tls is unset.                                                                  |
| `--tls-key-data`             | `GOALERT_TLS_KEY_DATA`             | Specifies a PEM-encoded private key. Has no effect if --listen-tls is unset.                                                                            |
| `--tls-key-file`             | `GOALERT_TLS_KEY_FILE`             | Specifies a path to a PEM-encoded private key file. Has no effect if --listen-tls is unset.                                                             |
| `--twilio-base-url`          | `GOALERT_TWILIO_BASE_URL`          | Override the Twilio API URL.                                                                                                                            |
| `--ui-dir`                   | `GOALERT_UI_DIR`                   | Serve UI assets from a local directory instead of from memory.                                                                                          |
| `--verbose`, `-v`            | `GOALERT_VERBOSE`                  | Enable verbose logging.                                                                                                                                 |
