# GoAlert

GoAlert is an on-call alerting platform written in Go.

## All-In-One (demo) Container

The quickest way to explore GoAlert is by using the GoAlert [all-in-one container](https://hub.docker.com/r/goalert/all-in-one).

- Ensure you have Docker Desktop installed ([Mac](https://docs.docker.com/docker-for-mac/release-notes/) / [Windows](https://docs.docker.com/docker-for-windows/release-notes/))
- `docker run -it --rm --name goalert-demo -p 8081:8081 goalert/all-in-one`

Using a web browser, navigate to `http://localhost:8081` and log in with user `admin` and password `admin123`.

## Development

Ensure you have docker, Go, node (and yarn), and make installed.

- If you do not have Postgres installed/configured, first run `make postgres`, GoAlert is built and tested against Postgres 11.
- For the first start, run `make regendb` to migrate and add test data into the DB. This includes an admin user `admin/admin123`.
- To start GoAlert in development mode run `make start`.
- To build the GoAlert binary run `make bin/goalert BUNDLE=1`.

### Automated Browser Tests

To run automated browser tests, you can start Cypress in one of the following modes:

- `make cy-wide` Widescreen format, in dev mode.
- `make cy-mobile` Mobile format, in dev mode.
- `make cy-wide-prod` Widescreen format, production build.
- `make cy-mobile-prod` Mobile format, production build.

### Running Smoketests

A suite of functional/behavioral tests are maintained for the backend code. These test various APIs and behaviors
of the GoAlert server component.

Run the full suite with `make smoketest`.

### Running Unit Tests

All unit tests can be run with `make test`.

UI Unit tests are found under the directory of the file being tested, with the same file name, appended with `.test.js`. They can be run independently with `make jest`. Watch mode can be enabled with `make jest JEST_ARGS=--watch`.

### Setup Postgres

By default, the development code expects a postgres server configured on port `5432`, with the user and DB `goalert`.

Alternatively, you can run `make postgres` to configure one in a docker container.

- You can reset the dev database with `make resetdb`
- You can reset and generate random data with `make regendb`, this includes generating an admin user `admin/admin123`
