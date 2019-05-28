# Contributing to GoAlert

We welcome feature requests, bug reports and contributions for code and documentation.

## Reporting Issues

Reporting bugs can be done in the GitHub [issue tracker](https://github.com/target/goalert/issues). Please search for a possible pre-existing issue first to help prevent duplicates.

Please include the version (`goalert version`) with new bug reports.

## Code Contribution

GoAlert is already used in production environments, so any new changes/features/functionality must (where possible):

- Not alter existing behavior without an explicit config change
- Co-exist with older versions without disruption
- Must have a safe way to disable/roll-back

It should always be safe to roll out a new version of GoAlert into an existing environment/deployment without downtime.

As an example, things like DB changes/migrations should preserve behavior across revisions.

## Pull Requests

Patches are welcome, but we ask that any significant change start as an [issue](https://github.com/target/goalert/issues/new) in the tracker, prefereably before work is started.

Be sure to run `make check` before opening a PR to catch common errors.

### UI Change Guidelines

- Complex logic should be broken out with corresponding unit tests (we use [Jest](https://jestjs.io/docs/en/using-matchers)) into the same directory. For example: [util.js](./web/src/app/rotations/util.js) and [util.test.js](./web/src/app/rotations/util.test.js).
- New functionality should have an integration test (we use [Cypress](https://docs.cypress.io/guides/getting-started/writing-your-first-test.html#Write-a-simple-test) for these) testing the happy-path at a minimum. Examples [here](./web/src/cypress/integration/sidebar.ts), and [more information here](./web/src/cypress/README.md).
- React components should follow React idioms, using common prop names, and having prop-types defined.

### Backend Change Guidelines

- Use unit tests as a tool to validate complex logic
- New functionality should have a behavioral smoketest at a minimum. For [example](./smoketest/simplenotification_test.go). Documentation on our smoketest framework can be found [here](./smoketest/README.md).
- New Go code should pass `golint`, exported functions/methods should be commented, etc..
