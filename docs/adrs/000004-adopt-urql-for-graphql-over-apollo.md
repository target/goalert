# Switching from Apollo to URQL for GraphQL on GoAlert

- Status: Accepted
- Date: 2022-06-01

**Acceptance Criteria**:

Clear Consensus: Simple Majority of requested reviewers AND no rejections

Technical Story: https://github.com/target/goalert/issues/3240

## Context and Problem Statement

The GoAlert project has been using Apollo for GraphQL management. However, with the evolution of React and the increasing complexity of Apollo's "magic" features, the development team has encountered challenges. How can we simplify our GraphQL client to better align with React hooks and follow the principle of "do one thing and do it well"?

## Considered Options

- Continue using Apollo for GraphQL
- Switch to URQL for GraphQL

## Decision Outcome

Chosen option: "Switch to URQL", because it aligns well with React hooks, providing a simpler and more predictable development experience. It adheres to the Unix philosophy of "do one thing and do it well," offering less "magic" and more transparent control over GraphQL data handling.

### Positive Consequences

- Better integration with the React ecosystem, leveraging hooks for state management.
- Simplification of the GraphQL client, reducing the learning curve for new developers.
- Enhanced transparency and control in data fetching and caching mechanisms.
- Reduction in package size and improved performance due to the lightweight nature of URQL.

### Negative Consequences

- The transition period may slow down current development as the team adapts to the new client.
- Possible need to refactor existing code that relies on Apollo-specific features.

## Pros and Cons of the Options

### Continue Using Apollo for GraphQL

- ✅ Good, because the team has existing experience with Apollo.
- ❌ Bad, because it may introduce unnecessary complexity with advanced features not required for the project.
- ❌ Bad, because it can obscure data management with its "magic" features.

### Switch to URQL for GraphQL

- ✅ Good, because URQL provides a simpler, more predictable approach consistent with React's core principles.
- ✅ Good, because it reduces the project's reliance on complex, feature-heavy libraries.
- ❌ Bad, because it requires the team to invest time in learning and migrating to a new GraphQL client.

## Links

- [URQL Documentation](https://formidable.com/open-source/urql/docs/)
- [Issue Link](https://github.com/target/goalert/issues/3240)
