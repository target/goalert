# Contributing to GoAlert

We welcome feature requests, bug reports and contributions for code and documentation.

## Reporting Issues

Reporting bugs can be done in the GitHub [issue tracker](https://github.com/target/goalert/issues). Please search for existing issues first to help prevent duplicates.

Please include the version (`goalert version`) with new bug reports.

## Code Contribution

GoAlert is already used in production environments, so any new changes/features/functionality must, where possible:

- Not alter existing behavior without an explicit config change
- Co-exist with older versions without disruption
- Must have a safe way to disable/roll-back

It should always be safe to roll out a new version of GoAlert into an existing environment/deployment without downtime.

As an example, things like DB changes/migrations should preserve behavior across revisions.

## Pull Requests

Patches are welcome, but we ask that any significant change start as an [issue](https://github.com/target/goalert/issues/new) in the tracker, preferably before work is started.

More information is available for [complex features](./docs/complex-features.md).

Be sure to run `make check` and tests before opening a PR to catch common errors.

### UI Change Guidelines

- Complex logic should be broken out with corresponding unit tests (we use [Jest](https://jestjs.io/docs/en/using-matchers)) into the same directory. For example: [util.js](./web/src/app/rotations/util.js) and [util.test.js](./web/src/app/rotations/util.test.js).
- New functionality should have an integration test (we use [Cypress](https://docs.cypress.io/guides/getting-started/writing-your-first-test.html#Write-a-simple-test) for these) testing the happy-path at a minimum. Examples [here](./web/src/cypress/integration/sidebar.ts), and [more information here](./web/src/cypress/README.md).
- React components should follow React idioms, using common prop names, and having prop-types defined.

### Backend Change Guidelines

- Use unit tests as a tool to validate complex logic. For [example](./schedule/rule/weekdayfilter_test.go).
- New functionality should have a behavioral smoke test at a minimum. For [example](./test/smoke/simplenotification_test.go). Documentation on our smoke test framework can be found [here](./test/smoke/README.md).
- Go code should [follow best practices](https://golang.org/doc/effective_go.html), exported functions/methods should be commented, etc..

## Testing

GoAlert utilizes 3 main types of testing as tools for different purposes:

- Unit tests are used for complicated logic and exhaustive edge-case testing and benchmarking. They live with the code being tested:
  - For backend code, a `_test.go` version of a file will contain relevant unit tests. More info [here](https://pkg.go.dev/testing)
  - For UI code, a `.test.ts` version of a file will contain relevant unit tests. More info [here](https://jestjs.io/docs/getting-started).
- Smoke tests (in `test/smoke`) are used to ensure main functionality and things like behavioral compatibility with future versions & DB migrations. These focus on hard guarantees like deliverability and preserving intent as the application and datastore evolves and changes over time.
- Integration tests (currently under `web/src/cypress/integration`) are primarily used to validate happy-path flows work end-to-end, and any important/common error scenarios. They are focused on UX and high-level functionality.
