# Development Setup

This guide assumes you have the commands `podman` (or `docker`), `go` (>= 1.21), `node` (>= 16.10), and `make` installed/available.

Targets like `make start` will automatically fallback to the `docker` command if `podman` is not available. The container tool command can be overridden by setting the `CONTAINER_TOOL` variable.

If you use Visual Studio Code, run `make vscode` to setup integrations for the project.

```bash
# force podman
make start CONTAINER_TOOL=podman

# force docker
make start CONTAINER_TOOL=docker
```

### Cross-Platform Images

To build cross-platform container images you will need `qemu-user-static` installed.

If using MacOS with `podman` you can do this with the following one-time commands:

```sh
podman machine ssh sudo rpm-ostree install qemu-user-static
podman machine ssh sudo systemctl reboot
```

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

### Manual Configuration

If you already have Postgres running locally you can create the `goalert` role.

```sql
CREATE ROLE goalert WITH LOGIN SUPERUSER;
```

Currently the dev user must be a superuser to enable `pgcrypto` with `CREATE EXTENSION`.

#### Toolchain Requirements

- For the first start, run `make regendb` to migrate and add test data into the DB (you can also scale the amount of random data with `SIZE` like `make regendb SIZE=10`). This includes adding an admin user `admin/admin123`.
- To start GoAlert in development mode run `make start`.
- To build the GoAlert binary run `make bin/goalert BUNDLE=1`.

## Automated Browser Tests

### Cypress Tests

To run automated browser tests, you can start Cypress in one of the following modes:

- `make cy-wide` Widescreen format, in dev mode.
- `make cy-mobile` Mobile format, in dev mode.
- `make cy-wide-prod` Widescreen format, production build.
- `make cy-mobile-prod` Mobile format, production build.

The Cypress UI should start automatically.

More information about browser tests can be found [here](../web/src/cypress/README.md).

### Playwright Tests

To run automated browser tests, you can start Playwright in one of the following modes:

- make playwright-ui Run all tests in UI mode.
- make playwright-run Run all tests in headless mode.

### Running Smoke Tests

A suite of functional/behavioral tests are maintained for the backend code. These test various APIs and behaviors
of the GoAlert server component.

Run the full suite with `make test-smoke`.

More information about smoke tests can be found [here](../test/smoke/README.md).

### Running Unit Tests

All unit tests can be run with `make test-unit`.

UI Unit tests are found under the directory of the file being tested, with the same file name, appended with `.test.js`. They can be run independently of the Go unit tests with `make jest`. Watch mode can be enabled with `make jest JEST_ARGS=--watch`.
