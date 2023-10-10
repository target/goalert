# Adoption of `sqlc` for Improved Query Management in GoAlert

- Status: accepted
- Date: 2022-06-01

**Acceptance Criteria**: Clear Consensus: Simple Majority of requested reviewers AND no rejections

Technical Story: 
- Proposal: [Migrate from util.Prepare to sqlc queries](https://github.com/target/goalert/issues/3108)
- Progress: [go: use sqlc for all db calls](https://github.com/target/goalert/issues/3235)

## Context and Problem Statement

In GoAlert, we manage our database queries using "stores", with SQL queries represented as strings directly in Go as prepared statements. This approach is harder to maintain and read, affecting the development experience. How can we enhance the readability, maintainability, and overall developer experience for our database queries?

## Decision Drivers

- Developer experience
- Code maintainability
- Readability of database queries
- Future scalability
- Efficient utilization of database resources
- Build-time guarantees

## Considered Options

- Continue using "stores" with strings directly in Go as prepared statements
- Adopt `sqlc` as a replacement for managing the majority of our DB queries
- Use GORM or another ORM/SQL query builder library

## Decision Outcome

Chosen option: "Adopt `sqlc` as a replacement for managing the majority of our DB queries", because it offers a more structured and readable way of managing SQL queries, improving the developer experience and overall maintainability.

### Positive Consequences

- Improved readability of database queries
- Enhanced developer experience due to better tooling/IDE support with `.sql` files
- Reduction in potential for SQL related errors
- Efficient utilization of database resources, preparing queries as-needed
- Future possibility of using PostgreSQL itself for build-time query validation

### Negative Consequences

- Initial learning curve for developers unfamiliar with `sqlc`
- Migration efforts required to transition from the current system to `sqlc`

## Pros and Cons of the Options

### Continue using "stores" with strings directly in Go as prepared statements

- Good, because it's the current system, and no changes would be required
- Bad, because it's harder to read and maintain
- Bad, because it immediately prepares and consumes database resources on startup
- Bad, because it doesn't scale well with the growth of the application

### Adopt `sqlc` as a replacement for managing the majority of our DB queries

- Good, because it offers a more structured and readable way of managing SQL queries
- Good, because `.sql` files provide enhanced tooling/IDE support
- Good, because it prepares queries as-needed, optimizing database resources
- Good, because of the potential for PostgreSQL-backed build-time query validation
- Bad, because there's an initial learning curve for developers unfamiliar with `sqlc`

### Use GORM or another ORM/SQL query builder library

- Good, because it could offer a different set of features and benefits
- Bad, because past experience with GORM led to complications, and it was replaced with `sqlc`
- Bad, because it would also have an associated learning curve and migration effort

## Links

- [sqlc GitHub repository](https://github.com/kyleconroy/sqlc)
- [sqlc documentation](https://sqlc.dev/)
