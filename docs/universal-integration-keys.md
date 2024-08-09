# Universal Integration Keys for GoAlert

## 1. Introduction

The universal integration keys (UIK) feature is a new addition to the GoAlert platform that allows users to configure ingress rules for any existing system capable of sending JSON data to a URL endpoint. This feature enables GoAlert to support new systems without direct integration by allowing users to configure rules to generate alerts based on the JSON payload. Additionally, UIKs introduce the ability to send non-critical messages (signals) in addition to creating alerts.

UIKs open the door for connecting GoAlert as a last-leg delivery to other systems without an intermediary layer. In a monitoring pipeline, it fits after the "there is something to report" stage and covers final filtering, delivery-preference, and data-mapping cases.

## 2. Motivation

Current integration keys in GoAlert have separate endpoints and hard-coded logic to map data structures from other systems to alerts. They only support creating alerts and do not support sending non-critical messages or signals. The UIK feature addresses these limitations by providing a flexible and programmable approach to handle incoming requests and generate appropriate actions.

## 3. Detailed Design

### 3.1 New Integration Key Type

The new UIK type will have the same HTTP request perspective as existing integration keys. Instead of fixed code mapping, the request body will be passed to an evaluation engine, and a list of actions will be generated based on user-defined rules.

Users will create a UIK in the same way as any other integration key, but after creation, they will land on the Rule Editor. The Rule Editor allows configuring all features of a UIK. It will list out all rules, and for each rule, it will display its condition expression and a summary of configured actions.

### 3.2 Signed Tokens

UIKs will use signed tokens for enhanced security, similar to sessions and calendar subscription tokens. The token will include a signature (ECDSA) and will be base64 encoded. From a user perspective, the only change will be a longer token string compared to existing integration keys.

### 3.3 Expr Language for Rules and Actions

The Expr language is chosen for its safety, performance, and Go integration capabilities. It provides a simple syntax for defining complex scenarios and ensures termination and memory safety. The full set of built-in functions for Expr will be available, with the addition of the `sprintf` function (which maps to the Go `Sprintf` function).

### 3.4 Data Mapping for Sent Messages

Data mapping is part of the action definition in UIKs. Helper functions like `sprintf` can be used to construct messages based on the input data. This allows users to customize the content of sent messages based on the incoming JSON payload.

There will be limits (TBD) on the number of output actions for a single key, total number of rules, and size of the Expr expression to enforce some level of bounding on request complexity. This feature will inherit the 1-request-per-process-per-key rule that other integration keys adopt, ensuring no "noisy neighbor" issues with isolation.

### 3.5 Max Pending Messages Per Service

Instead of de-duplication, the system will enforce a maximum number of pending messages per service and per service per destination. If the number of pending messages exceeds these limits, any further requests resulting in that destination will be rejected with a 429 status code.

### 3.6 Immediate Handling for Specific Destinations

Webhook destinations and create-alert destinations will be handled immediately and inline with the originating request. Other types of destinations will be queued for processing.

### 3.7 Queued Messages

Queued messages will be subject to the existing GoAlert prioritization and rate limiting protections. This ensures that all queued messages are managed efficiently and in accordance with system policies.

### 3.8 Auth and Key Rotation

For authentication, keys will support rotation. This means users can perform a primary/secondary token swap to rotate keys without downtime.

### 3.9 Single Message Per Destination Rule

A rule or set of actions may only result in a single message per destination, ensuring clarity and preventing message overload.

## 4. API Changes

API documentation will be updated to include the new integration key type, though since the request body is arbitrary, it will be brief and just indicate the URL path and guidance on how to create a key.

## 5. Backward Compatibility

There will be no impact or changes to any existing APIs or functionality.

## 6. Security Considerations

The use of signed tokens (ECDSA) enhances the security of UIKs compared to the simple UUIDs used in existing integration keys.

Existing rate limits (e.g., messages to a Slack channel) will be observed, as well as prioritization. Alert notifications, for example, will always take priority when necessary to signal notifications. While Expr ensures always-terminating, there will also be a max execution time enforced per-request. A single user may only have 1 request active per key, and each key will have its critical execution limited to 1 second.

## 7. Performance and Scalability

The Expr language is optimized for speed, utilizing an optimizing compiler and a bytecode virtual machine. In practice, even extremely complex configurations (with many rules) execute in fractions of a millisecond, and average use cases execute in microseconds.

Preliminary benchmark results:

While a bit arbitrary, the following benchmarks were run as a sanity test for some worst-case scenarios.

```
cpu: Intel(R) Core(TM) i9-14900K

BenchmarkMostActions-32  6034     189660 ns/op    102333 B/op    929 allocs/op
BenchmarkMostRules-32     715    1754745 ns/op    870548 B/op   3590 allocs/op
```

The compiled bytecode will be cached in an LRU cache, and the limitations on concurrency, along with code size constraints, will bound memory use. Lastly, the number of integration keys per service is already bounded, ensuring a total bounded limit to resource consumption per user/service.

## 8. Future Enhancements

The possibility of extending UIKs to replace the code in existing integration keys with pre-set configurations that can be tweaked by users is mentioned as a potential future enhancement.
