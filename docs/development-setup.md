# Development Setup

This guide assumes you have the following commands installed/available:

- `docker`
- `go` (>= 1.21)
- `node` (>= 18)
- `make`

If you are using vscode, you can run `make vscode` to fix the workspace settings for use with the UI code under `web/src`.

## Quick Start

To start the development environment, simply run:

```bash
make start
```

You can then access the GoAlert UI at [http://localhost:3030](http://localhost:3030) and login with the default credentials `admin/admin123`.

In dev mode there is a `Dev` item in the navbar that allows you to configure and access some additional integrations like Prometheus and Email messages.

## External Traffic

To do local development with external traffic you will need a publicly-routable URL and can start localdev with `PUBLIC_URL` set. For example:

```bash
make start PUBLIC_URL=http://localdev.example.com
```

You may add additional startup commands to the `Procfile.local` file to have them automatically run with `make start` and similar commands.

```bash
ngrok: ngrok http -subdomain=localdev 3030
```

## Database (PostgreSQL)

GoAlert is built and tested against Postgres 13. Version 11+ should still work as of this writing.

The easiest way to setup Postgres for development is to run `make postgres`.
You can connect to the local DB at `postgres://goalert@localhost:5432/goalert` (no password).
This will start a container with the correct configuration for the dev environment.

### Database Management

To reset or regenerate the database (e.g., to resolve migration errors or test with a different dataset size), run:

```bash
make regendb
```

You can scale the amount of random data with the `SIZE` parameter. For example:

```bash
make regendb SIZE=10
```

This command also includes adding an admin user with the credentials `admin/admin123`.

#### Manual Database Configuration

If you already have Postgres running locally you can create the `goalert` role.

```sql
CREATE ROLE goalert WITH LOGIN SUPERUSER;
```

Currently the dev user must be a superuser to enable `pgcrypto` with `CREATE EXTENSION`.

### Cypress Tests

To run automated browser tests, you can start Cypress in one of the following modes:

- `make cy-wide` Start Cypress UI in widescreen format, dev mode.
- `make cy-mobile` Start Cypress UI in mobile format, dev mode.
- `make cy-wide-prod` Start Cypress UI in widescreen format, production build.
- `make cy-mobile-prod` Start Cypress UI in mobile format, production build.
- `make cy-wide-prod-run` Run tests in headless mode, widescreen format, production build.
- `make cy-mobile-prod-run` Run tests in headless mode, mobile format, production build.

The Cypress UI should start automatically.

More information about browser tests can be found [here](../web/src/cypress/README.md).

### Playwright Tests

To run automated browser tests, you can start Playwright in one of the following modes:

- `make playwright-ui` Start the Playwright UI.
- `make playwright-run` Run all tests in headless mode.

### Running Smoke Tests

A suite of functional/behavioral tests are maintained for the backend code. These test various APIs and behaviors
of the GoAlert server component.

Run the full suite with `make test-smoke`.

More information about smoke tests can be found [here](../test/smoke/README.md).

### Running Unit Tests

All unit tests can be run with `make test-unit`.

UI Unit tests are found under the directory of the file being tested, with the same file name, appended with `.test.js`. They can be run independently of the Go unit tests with `make jest`. Watch mode can be enabled with `make jest JEST_ARGS=--watch`.
