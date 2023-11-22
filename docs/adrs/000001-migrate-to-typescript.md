# Migration from JavaScript to TypeScript for GoAlert

- Status: Accepted
- Date: 2022-04-19

**Acceptance Criteria**: Clear Consensus: Simple Majority of requested reviewers AND no rejections

Technical Story: https://github.com/target/goalert/issues/2318

## Context and Problem Statement

The existing JavaScript codebase has become challenging to maintain and scale. The lack of static typing leads to common runtime errors and hinders the development speed. How can we improve the developer experience and enhance code quality for GoAlert?

## Considered Options

- Maintain the current JavaScript codebase
- Migrate to TypeScript

## Decision Outcome

Chosen option: "Migrate to TypeScript", because it introduces static typing to the codebase, which enhances code quality, reduces the potential for runtime errors, and improves integration with the backend schema.

### Positive Consequences

- Improved code quality through static type checking, reducing runtime errors.
- Enhanced developer experience with better tooling support in IDEs, facilitating easier coding and debugging.
- Streamlined code maintenance and review processes, leading to more reliable code.
- Improved integration with backend schema, ensuring that frontend and backend data structures are in sync.

### Negative Consequences

- Increased initial development time due to learning curve and setup of TypeScript.
- Potential need for initial rewriting of existing JavaScript code to conform to TypeScript's type system.

## Pros and Cons of the Options

### Maintain Current JavaScript Codebase

- ✅ Good, because the team is already familiar with JavaScript.
- ❌ Bad, because it leads to less maintainable code and potential for bugs.
- ❌ Bad, because it lacks the advanced features and tooling that TypeScript provides.

### Migrate to TypeScript

- ✅ Good, because TypeScript's static typing system improves code reliability and maintainability.
- ✅ Good, because it has better tooling and IDE support for a smoother development process.
- ✅ Good, because it aligns with modern development practices and can help in future-proofing the codebase.
- ❌ Bad, because of the initial investment in time and resources to migrate and train the team.

## Links

- [Issue Link](https://github.com/target/goalert/issues/2318)
- [TypeScript Official Documentation](https://www.typescriptlang.org/docs/)
