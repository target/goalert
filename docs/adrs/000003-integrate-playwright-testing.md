# Migration from Cypress to Playwright for GoAlert Testing

- Status: accepted
- Date: 2022-09-13

**Acceptance Criteria**: Clear Consensus

Technical Story: Due to several limitations with Cypress, such as flaky selector chains and difficulty in testing popup links and oauth flows, it was decided to migrate to Playwright for a more robust testing environment.

## Context and Problem Statement

We aim to address the shortcomings of Cypress used in GoAlert, such as its flaky selector chains, the inability to test popup links and separate servers/hosts, and the frequent breaking changes with updates. How can we improve the test infrastructure to be more reliable and maintainable?

## Decision Drivers

- Cypress's limitations in handling complex test scenarios
- The need for better integration with existing IDEs
- Requirement for concurrent tests and multi-tab/window testing capabilities

## Considered Options

- Continuing with the current setup and adding workarounds for Cypress's limitations
- Migrating to Playwright for testing

## Decision Outcome

Chosen option: "Migrate to Playwright", because it offers better integration with existing IDEs, allows for more flexible testing capabilities such as handling multiple tabs/windows, and supports concurrent test execution.

### Positive Consequences

- Improved test reliability and maintainability
- Enhanced ability to test complex user interactions and flows
- Reduction in time and maintenance cost due to less frequent breaking changes

### Negative Consequences

- Initial setup and migration effort
- Learning curve associated with adopting a new testing framework

## Pros and Cons of the Options

### Continuing with Cypress

- ✅ Good, because it is already in place and the team is familiar with it.
- ❌ Bad, because it introduces complexity and maintenance challenges.
- ❌ Bad, because it does not support certain types of tests, leading to limited coverage.

### Migrating to Playwright

- ✅ Good, because it resolves existing issues with Cypress.
- ✅ Good, because it allows for more robust and comprehensive testing.
- ❌ Bad, because of the initial effort required to set up and learn the new system.

## Links

- [Relevant Issue](https://github.com/target/goalert/issues/2589)
- [Migration Commit](https://github.com/target/goalert/pull/2608)
- [Playwright Documentation](https://playwright.dev/)
