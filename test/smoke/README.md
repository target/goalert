# Behavioral Smoketest Suite

A suite of tests to check that notifications get sent properly.

## Setup

1. Ensure you have postgres running locally, the test suite will create timestamped databases while running.

## Running Tests

Run `make smoketest` from the root of the repo to run all tests.
The script will automatically rebuild goalert, including any change migrations.

To run a single test you can use `go test` from the smoketest directory.

### Experimental Flags

Tests that need an experimental flag pass it explicitly via `harness.NewHarnessWithFlags`.

To run the whole suite (or a subset) with additional experimental flags enabled, without editing any tests, set `SMOKE_EXP_FLAGS` to a comma-separated list of flag names (the same names accepted by `goalert --experimental`):

```sh
cd test/smoke && SMOKE_EXP_FLAGS=example,univ-keys go test -run 'Rotation|Schedule' .
```

The flags are merged into the set used by every harness, including those created with `NewHarness` and `NewHarnessWithFlags`. This is how a change gated behind an experimental flag is proven to preserve behavior: the unchanged smoke tests must pass with the flag off (the default) and with it on.

## Creating Tests

- Try to keep the test under 1 minute, where possible.
- Use the latest migration when the test is created.
- Make sure to call `t.Parallel()` and `defer h.Close()` in your test.

## Engine Cycles and River Jobs

The harness drives the backend engine explicitly rather than waiting for its timer: `h.Trigger()` runs one engine cycle and waits for it to finish. Most assertion helpers (`ExpectSMS`, Slack `ExpectMessage`, `WaitAndAssertOnCallUsers`, ...) call `Trigger` for you inside their retry loops.

Some engine modules do their work in [River](https://riverqueue.com) jobs instead of inline during the cycle (the `rotation-manager` queue today; `schedule-manager` is planned). Those jobs are enqueued by DB triggers and River's scheduler exactly as in production; the harness never fakes or re-inserts them. Instead, `Trigger` waits for the engine queues to drain before and after the cycle, and runs one extra cycle if it had to wait so modules that consume job output (e.g. the escalation manager) see it. Tests with no job activity get exactly one cycle per `Trigger`.

The following helpers are available for tests that interact with jobs directly:

- `h.WaitForJobs(queues ...string) bool` waits (up to 30s) until the given queues have no available, running, pending, or due (`scheduled`/`retryable` with `scheduled_at` in the past, by mocked DB time) jobs. It returns `true` if it had to wait at all.
- `h.CountJobs(queue, kind string, argsContains map[string]any, states ...string) int` counts jobs in a queue by kind, JSON-containment match on `args` (`@>`), and (optionally) state.
- `h.EnqueueJob(queue, kind, id string, args map[string]any)` inserts a job via `fn_util_river_job` (the same function DB triggers use) and then waits for the queue to drain.

Note that `h.FastForward` only moves the DB clock; jobs scheduled for the new time are picked up by River's scheduler within a few seconds, and `Trigger`/`WaitForJobs` account for that delay.
