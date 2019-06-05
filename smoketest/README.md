# Behavioral Smoketest Suite

A suite of tests to check that notifications get sent properly.

## Setup

1. Ensure you have postgres running locally, the test suite will create timestamped databases while running.
1. Make sure you have the `goalert` binary installed and up-to-date. (run `make install` from the root of the repo)

## Running Tests

Run `make smoketest` from the root of the repo to run all tests.
The script will automatically rebuild goalert, including any change migrations.

To run a single test you can use `go test` from the smoketest directory.

## Creating Tests

- Try to keep the test under 1 minute, where possible.
- Use the latest migration when the test is created.
- Make sure to call `t.Parallel()` and `defer h.Close()` in your test.
