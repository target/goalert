# API Key Architecture for System Integrations

- Status: accepted
- Date: 2023-07-20

**Acceptance Criteria**:

- Clear Consensus: Simple Majority of requested reviewers AND no rejections

Technical Story: #3007

## Context and Problem Statement

With the need for secure system integrations, we evaluated the best approach to provide API access. We require a method that leverages our existing GraphQL infrastructure, maintains our security posture, and avoids the complexity of substantial refactoring or maintaining parallel API systems.

## Decision Drivers

- Security: Ensuring that the API keys cannot be easily exploited.
- Maintainability: Avoiding the need for extensive refactoring and simplifying ongoing maintenance.
- Accessibility: Ensuring keys can be managed and used effectively without complex overhead.

## Considered Options

- Extensive use of gRPC for integrations
- Implementing a scope-type system for GraphQL
- Creating a separate REST API
- Unofficial use of user session tokens for API access
- Restricted API keys with predefined queries and operations

## Decision Outcome

Chosen option: "Restricted API keys with predefined queries and operations", accepted in July 2023, because it allows for secure and targeted use cases without the need for extensive system overhaul or the security risks associated with broader access.

### Positive Consequences

- Aligns with the existing permission and authorization infrastructure.
- Provides a secure way to integrate systems with the ability to fine-tune access.
- Avoids the complications and overhead associated with gRPC, scope systems, or REST APIs.

### Negative Consequences

- Manual key rotation could become a management challenge as the number of integrations grows.
- The absence of inbuilt alerting mechanisms for key usage may necessitate the development of external monitoring solutions.

## Pros and Cons of the Options

### Extensive use of gRPC for integrations

- ❌ Bad, because it does not work well with load balancers and requires careful setup.

### Implementing a scope-type system for GraphQL

- ❌ Bad, because it requires substantial refactoring and ongoing maintenance of scopes.

### Creating a separate REST API

- ❌ Bad, because it would duplicate existing functionality and increase maintenance burden.

### Unofficial use of user session tokens for API access

- ❌ Bad, because it is not secure and not officially supported.

### Restricted API keys with predefined queries and operations

- ✅ Good, because it aligns with existing authorization checks and infrastructure.
- ✅ Good, because it offers a secure and scalable solution for system integrations.
- ✅ Good, because it avoids the need for a significant system overhaul.

## Links

- [Feature Request](https://github.com/target/goalert/issues/3007)


## Implementation Details and Reasoning

The architecture of the API key system is designed with a focus on security and specificity. Below are some of the prominent implementation details and the reasoning behind these choices:

- **Fixed Query with Embedded Hash**: When generating an API key, a specific GraphQL query document is required. The key will only be functional with that exact query document. This decision was made to restrict access to predetermined operations, enhancing security by preventing arbitrary query execution. A SHA256 hash of this query is embedded in the generated JWT token, ensuring the integrity of the key and that it cannot be altered without detection.

- **Admin-Only Access**: API keys are exclusively available to admins. This decision simplifies the system by avoiding the complexity of implementing organizational-level policies, such as maximum expiration time, at the initial stages. It also focuses the use of API keys on system integrations rather than individual use, which aligns with the current needs.

- **Mandatory Expiration Time**: All API keys must have an expiration time, although the admin can decide the duration. By enforcing a mandatory expiration, we ensure that keys cannot be left indefinitely, which reduces the risk of security breaches over time.

- **Usage Metrics**: The system records the last usage of an API key up to once per minute. This includes the key ID, IP address, and reported user agent, providing essential information for monitoring and auditing purposes.

- **Role-Based Access**: API keys can be assigned a user or admin role. The current implementation supports keys with anonymous roles, meaning they do not impersonate users and can only interact with non-user-specific functions. This decision aids in maintaining clear audit trails and reduces the risk associated with user impersonation.

- **Manual Rotation**: The rotation of API keys is a manual process. When a key is created, it can be seen or copied only once, and it is the responsibility of the admin to securely transfer the key to the intended system. While this may be seen as a limitation, it is a deliberate choice to avoid complexity and automate security-sensitive operations at this stage.

- **Duplication Feature**: To aid in key management, a "Duplicate" function is available in the UI, which copies the parameters of an existing key into a new one. This feature is designed to simplify the process of rotating keys by pre-filling information based on existing keys.

These decisions form the core of our API key architecture, prioritizing security, manageability, and alignment with our existing systems. The chosen approach balances the need for secure integration with the practical aspects of implementation and management within the current GoAlert infrastructure.
