## Development Setup

#### [PostgreSQL](https://www.postgresql.org/) Requirements (Version 11 recommended)

- A dedicated database (i.e. goalert)
- A user (role) with appropriate (owner) permissions on the database (i.e. goalertuser)
- Enable required extension [pgcrypto](https://www.postgresql.org/docs/11/pgcrypto.html)

The following is an example of how you may perform the above using native psql statements:

```
CREATE DATABASE goalert;
CREATE ROLE goalertuser WITH LOGIN PASSWORD '<complex_password>';
GRANT ALL PRIVILEGES ON DATABASE goalert TO goalertuser;
```

Change database to the newly created database `goalert` and enable `pgcrypto` extension:

```
CREATE EXTENSION pgcrypto;
```

#### Toolchain Requirements

Ensure you have Docker, Go, node (and yarn), and make installed.

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

By default, the development code expects a Postgres server configured on port `5432`, with the user and DB `goalert`.

Alternatively, you can run `make postgres` to configure one in a docker container.

- You can reset the dev database with `make resetdb`
- You can reset and generate random data with `make regendb`, this includes generating an admin user `admin/admin123`
