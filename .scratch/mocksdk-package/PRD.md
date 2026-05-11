# PRD: Mock SDK Package

Status: done
Created: 2026-05-08

## Problem Statement

MockServer already provides a runtime service and admin APIs for managing namespaces, rulesets, rules, publishing, simulation, snapshots, and rollback. The next part of the mock tool chain needs a reusable SDK package that `mockinject` can call from injected business-code hooks.

Today, a `mockinject` implementation would need to understand too much MockServer detail: how to convert an original HTTP request into the MockServer event shape, which mockserver endpoint to call, how to decode the response envelope, how to distinguish a matched mock response from a ruleset miss or rule miss, and what to do when namespace fallback says the original request should be forwarded. That coupling would make every mockinject implementation duplicate the same logic and would make fallback behavior hard to keep consistent.

The existing runtime HTTP endpoint is not enough as the SDK contract. It directly returns a mock response or executes namespace forward fallback on the MockServer side. The SDK use case is different: `mockinject` needs a decision that says either "return this response" or "forward the original request yourself" so that injected code can control the business process locally.

The existing published simulation endpoint is also not enough. It can explain published matching, but it does not resolve namespace fallback into a client-side decision protocol.

The user wants an SDK package that becomes the stable integration point between `mockinject` and MockServer.

## Solution

Build a public Mock SDK package and the minimum MockServer decision API needed to support it.

From the user's perspective, injected code should be able to pass the original HTTP request to the SDK and receive a simple typed decision:

- Return a mock response.
- Forward the original request.

The SDK should hide the MockServer event shape, response envelope decoding, environment override handling, timeout handling, trace propagation, and decision parsing. It should preserve the original request data required for forwarding, especially method, scheme, host, path, query, headers, and body.

MockServer should expose a runtime-safe decision endpoint that evaluates published rulesets and namespace fallback without executing client-side forward fallback on the server. The endpoint should return a stable decision contract for the SDK. It should reuse the same matching semantics as the published runtime path: published snapshots only, namespace-aware ruleset selection, selector checks, rule matching, action execution for response-producing mock actions, and namespace fallback resolution for ruleset miss and rule miss.

Environment variables should have the highest priority for special SDK configuration such as namespace ID and mockserver host. This allows CI, local scripts, and generated injected code to redirect behavior without changing business code or generated mockinject glue.

## User Stories

1. As a mockinject author, I want to call a single SDK function with an original HTTP request, so that injected code does not need to know MockServer API details.
2. As a mockinject author, I want the SDK to normalize an HTTP request into MockServer's event protocol, so that request matching is consistent with runtime behavior.
3. As a mockinject author, I want the SDK to return a typed mock response decision, so that injected code can write the mocked status, headers, and body back to the business process.
4. As a mockinject author, I want the SDK to return a typed forward decision, so that injected code can continue the original request when no mock should intercept it.
5. As a mockinject author, I want the SDK to preserve the original request body after normalization, so that forwarding still works after the SDK has inspected the request.
6. As a mockinject author, I want the SDK to preserve the original method, scheme, host, path, query, and headers, so that forwarding behavior remains faithful to the business request.
7. As a mockinject author, I want the SDK to propagate trace IDs, so that mock decisions can be correlated across business logs and MockServer logs.
8. As a mockinject author, I want the SDK to accept explicit configuration options, so that generated injected code can set host, namespace, timeout, and HTTP client behavior.
9. As a mockinject author, I want environment variables to override explicit SDK options, so that runtime environments can redirect mockserver host or namespace without rebuilding injected code.
10. As a mockinject author, I want clear SDK errors for missing mockserver host or invalid namespace, so that setup problems fail with actionable messages.
11. As a mockinject author, I want transport errors to be distinguishable from valid forward decisions, so that injected code does not accidentally treat MockServer downtime as a mock miss.
12. As a mockinject author, I want the SDK to expose the matched ruleset ID and rule ID when a mock hits, so that logs can explain which rule controlled the response.
13. As a mockinject author, I want the SDK to expose fallback reason when forwarding is requested, so that logs can distinguish ruleset miss from rule miss.
14. As a mockinject author, I want the SDK to expose response headers exactly enough for business clients, so that mocked responses behave like real upstream responses.
15. As a mockinject author, I want the SDK to avoid leaking admin concepts, so that injected runtime code only depends on runtime-safe APIs.
16. As a mockinject author, I want the SDK to avoid importing MockServer internal packages, so that it can be used as a real public package by external modules.
17. As a mockinject author, I want the SDK API to be small and stable, so that different mockinject implementations can share it without frequent rewrites.
18. As a mockinject author, I want a helper that converts an SDK response decision into an HTTP response shape, so that common injected HTTP-client adapters are easy to implement.
19. As a mockinject author, I want a low-level decision client as well as high-level helpers, so that unusual injected clients can customize execution.
20. As a mockinject author, I want the SDK to support context cancellation, so that injected requests do not hang after the business request is canceled.
21. As a mockinject author, I want SDK timeouts to be configurable, so that slow MockServer calls do not block business code indefinitely.
22. As a mockinject author, I want no automatic retry by default, so that the SDK does not duplicate non-idempotent decision calls or hide latency problems.
23. As a mockinject author, I want the SDK to allow an injected custom HTTP client, so that tests and enterprise network settings can be handled cleanly.
24. As a mockinject author, I want the SDK to include a test-friendly client interface, so that mockinject adapters can be unit tested without starting MockServer.
25. As a mockinject author, I want JSON response bodies to be handled without losing raw bytes, so that injected clients can preserve content semantics.
26. As a mockinject author, I want non-JSON response bodies to be handled safely, so that mocks can return plain text or binary-like payloads when needed.
27. As a mockinject author, I want hop-by-hop request headers to be handled predictably by helper forwarding code, so that generated adapters do not forward invalid HTTP headers.
28. As a MockServer user, I want namespace ID to be overridable from the environment, so that the same injected binary can target different mock namespaces in different runs.
29. As a MockServer user, I want mockserver host to be overridable from the environment, so that local, CI, and shared MockServer deployments can be switched without code changes.
30. As a MockServer user, I want the SDK decision to use published snapshots only, so that draft changes do not affect business runtime until they are published.
31. As a MockServer user, I want a ruleset miss default to become a forward decision when the namespace fallback is configured as forward, so that unmatched traffic passes through by default.
32. As a MockServer user, I want a rule miss default to become a forward decision when the namespace fallback is configured as forward, so that selected rulesets can still pass through unmatched rules.
33. As a MockServer user, I want response fallback to become a response decision, so that strict namespaces can intentionally block or stub unmatched traffic.
34. As a MockServer user, I want SDK decisions to use the same selector and rule matching semantics as runtime requests, so that UI simulation, runtime behavior, and injected behavior agree.
35. As a MockServer user, I want SDK decisions to include enough diagnostic metadata for logging, so that I can understand why traffic was mocked or forwarded.
36. As a MockServer user, I want SDK integration docs and examples, so that I can quickly wire a mockinject prototype to MockServer.
37. As a MockServer maintainer, I want the backend decision logic to be shared with runtime matching where appropriate, so that SDK behavior does not fork from runtime behavior.
38. As a MockServer maintainer, I want forward fallback resolution separated from forward fallback execution, so that runtime can execute server-side forwarding while SDK can return client-side forward decisions.
39. As a MockServer maintainer, I want a stable public decision contract, so that later SDKs for other languages can reuse the same protocol.
40. As a MockServer maintainer, I want SDK models to be independent from internal business objects, so that refactors inside MockServer do not break external users.
41. As a MockServer maintainer, I want backend tests to prove matched response, ruleset miss, rule miss, response fallback, and forward fallback decisions, so that default namespace behavior remains safe.
42. As a MockServer maintainer, I want SDK unit tests to cover normalization, config resolution, decision decoding, body replay, and error classification, so that SDK behavior is reliable without requiring end-to-end tests for every case.
43. As a MockServer maintainer, I want integration tests with an in-memory HTTP server, so that the SDK can be validated against real response envelopes and decision payloads.
44. As a MockServer maintainer, I want generated docs to state the environment override precedence explicitly, so that users do not misconfigure namespace or mockserver host.
45. As a MockServer maintainer, I want non-HTTP protocol support kept out of the first SDK package, so that the first version can be focused and correct for the current runtime surface.

## Implementation Decisions

- Implement a public Go SDK package as the first SDK target because the repository and current service are Go-based.
- Keep the SDK package outside internal-only package boundaries so external mockinject modules can import it.
- Do not expose MockServer internal business objects as the SDK's public contract. The SDK should define stable request, response, decision, config, and error types that can evolve independently.
- Define the SDK's primary user-facing concept as a `Decision`, with two required decision kinds: mock response and forward original request.
- Treat "mock response" as a decision that contains HTTP status, headers, body bytes or decoded body representation, matched state, ruleset ID, rule ID, and trace metadata.
- Treat "forward original request" as a decision that contains fallback reason, optional forward timeout, original request metadata required by the injector, and trace metadata.
- Add a runtime-safe MockServer decision endpoint for SDK usage. This endpoint should evaluate published rulesets and namespace fallback but should not perform client-side forward fallback on the server.
- Keep the existing runtime HTTP endpoint behavior unchanged. It may continue to execute server-side fallback forward because it is a direct runtime mock endpoint, not the SDK decision protocol.
- Keep the existing published simulation endpoint behavior unchanged. It remains a diagnostic and explain endpoint, not the SDK decision endpoint.
- Share matching semantics between runtime, simulation, and SDK decision generation. The SDK decision path must not invent a separate selector or rule matching model.
- Extract or introduce a backend decision module that encapsulates published ruleset matching plus namespace fallback resolution behind a small interface.
- Split fallback resolution from fallback execution. Runtime can resolve and then execute forward fallback, while SDK decision can resolve and return a forward decision.
- Preserve namespace semantics: default namespace exists lazily, missing non-default namespaces are errors, ruleset miss uses `ruleset_miss_action`, and rule miss uses `rule_miss_action`.
- Preserve default namespace fallback behavior: both ruleset miss and rule miss default to forwarding the original request with the default forward timeout.
- Use environment variables as the highest-priority SDK configuration source. At minimum, standardize variables for mockserver host and namespace ID.
- Use explicit SDK options as the next-priority configuration source after environment variables.
- Use safe SDK defaults only where they do not hide configuration mistakes. For example, timeout may have a default, but mockserver host should be explicit through options or environment.
- Make namespace ID resolution explicit: environment override first, SDK option second, request-derived or SDK default last.
- Make the SDK's HTTP request normalizer a deep module. It should own body reading, body restoration, method normalization, scheme detection, host and original host handling, query copying, header copying, client IP extraction, and trace ID propagation.
- Make the SDK's config resolver a deep module. It should own environment lookup, option merging, defaulting, validation, and final immutable client configuration.
- Make the SDK's decision client a deep module. It should own endpoint URL construction, request serialization, response envelope decoding, error classification, and timeout/context handling.
- Make the SDK's decision parser a deep module. It should turn raw decision API responses into stable typed SDK decisions and reject ambiguous or unsupported decision shapes.
- Provide high-level HTTP helpers on top of the low-level decision client, but keep them optional so mockinject implementations can control their own transport details.
- Do not automatically forward the original request inside the core SDK decision method. The core responsibility is to ask MockServer for a decision and return it to mockinject.
- If helper forwarding is provided, keep it separate from decision retrieval so callers can choose whether to use it.
- Do not add automatic retries by default. A retry policy can be added later as an explicit option if real usage requires it.
- Ensure request bodies can be consumed by SDK normalization without making the original request unusable for forwarding.
- Ensure SDK response body handling can represent JSON, text, and opaque bytes without forcing all users through a JSON-only path.
- Ensure backend decision responses carry enough metadata for observability: matched flag, fallback flag, fallback reason, ruleset ID, rule ID, and trace ID.
- Ensure backend decision responses are versionable or otherwise stable enough for future non-Go SDKs.
- Keep admin authentication separate from SDK decision calls. The decision endpoint should follow runtime behavior rather than admin management behavior unless a future runtime authentication feature is introduced.
- Keep mockinject itself out of scope. This PRD only defines the SDK package and any MockServer API needed by that SDK.
- Keep current frontend management UI out of scope except for documentation links or future references.
- Update user-facing documentation with the SDK workflow, environment variables, decision meanings, and a minimal example.

## Testing Decisions

- Good tests should validate external behavior and stable contracts, not private implementation details.
- Backend decision tests should cover a published ruleset hit returning a mock response decision.
- Backend decision tests should cover ruleset miss returning a forward decision when the namespace ruleset miss fallback is forward.
- Backend decision tests should cover rule miss returning a forward decision when the namespace rule miss fallback is forward.
- Backend decision tests should cover ruleset miss and rule miss returning response decisions when namespace fallback is response.
- Backend decision tests should cover the default namespace creating or resolving forward fallback consistently.
- Backend decision tests should cover missing non-default namespace errors.
- Backend decision tests should cover trace metadata propagation into decision responses.
- Backend controller tests should verify the SDK decision endpoint response envelope and HTTP status behavior.
- Backend tests should verify that the existing runtime endpoint still returns or forwards responses as before.
- SDK config resolver tests should cover environment variable precedence over explicit options.
- SDK config resolver tests should cover missing mockserver host, invalid host, invalid namespace, and timeout defaulting.
- SDK normalizer tests should cover method, scheme, host, original host, path, query, headers, raw body, parsed body, client IP, and trace ID mapping.
- SDK normalizer tests should prove that reading the request body does not prevent later forwarding.
- SDK decision client tests should use an in-memory HTTP server to verify endpoint construction, envelope decoding, success decisions, server errors, malformed JSON, and context cancellation.
- SDK decision parser tests should cover mock response decisions, forward decisions, fallback metadata, unknown decision kinds, and missing required fields.
- SDK helper tests should cover writing a response decision to an HTTP response target without dropping headers or status.
- SDK helper tests should cover helper forwarding only if helper forwarding is implemented in the first pass.
- Integration tests should run the SDK against a real in-process MockServer handler when feasible, covering at least one hit and one forward decision.
- Prior art exists in the repo's controller tests for published runtime flow, namespace fallback response flow, default forward fallback, rule miss forward fallback, admin auth, and MySQL-backed runtime integration.
- The final implementation should pass `go test ./...`.
- If documentation or examples include runnable commands, they should be smoke-checked manually or covered by tests where practical.

## Out of Scope

- Implementing mockinject or any compile-time code injection mechanism.
- Implementing SDKs for languages other than Go.
- Supporting non-HTTP protocols in the first SDK package.
- Changing ruleset, rule, selector, priority, publish, snapshot, or rollback semantics.
- Changing the existing runtime endpoint's server-side response and forward behavior.
- Replacing the existing admin simulation endpoints.
- Adding a new permissions, token, or authentication system for runtime/SDK calls.
- Adding a frontend SDK configuration page.
- Building a full CLI around the SDK.
- Adding persistent audit tables for SDK decisions.
- Adding production distributed tracing integrations beyond trace ID propagation.
- Implementing automatic retry policies by default.
- Guaranteeing binary body transformation fidelity beyond a safe first HTTP SDK representation, unless implementation discovers that the current body model is insufficient.

## Further Notes

- The rough SDK note defines the SDK as the shared layer used by mockinject to send original requests to MockServer and receive a decision.
- The mock tool chain is: MockServer manages runtime state, Mock SDK asks MockServer for a decision, and mockinject performs compile-time injection and executes the local behavior.
- The key product distinction is that MockServer's direct runtime endpoint can execute behavior, while the SDK path should return a decision for injected code to execute.
- The current backend already has most of the matching engine and namespace fallback data model required for this feature. The main missing piece is a decision contract that resolves fallback without forcing server-side forwarding.
- This project is greenfield, so the SDK package can choose the clean public API shape without preserving any previous SDK compatibility.

## Comments

- Implemented all seven Mock SDK package issues.
- Final SDK scope includes published hit response decisions, forward decisions for ruleset/rule miss, HTTP request normalization with body replay, environment-configured client setup, response fallback decisions, optional HTTP helpers, and user-facing quickstart docs.
- Verified with `go test ./...` and `git diff --check`.
