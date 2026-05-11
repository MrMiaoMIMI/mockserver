# PRD: Decision-Only Forward Boundary

Status: done
Created: 2026-05-08

## Problem Statement

MockServer is intended to be a generic mock tool chain component, not a protocol-specific proxy implementation. The current HTTP-focused implementation already has a useful decision model, but the Go Mock SDK also exposes concrete original-request forwarding helpers, and the backend runtime fallback implementation is close to the generic service layer.

That boundary is too heavy for the next stage of the project. Future protocol adapters may need to mock RPC systems such as SPEX or gRPC, cache clients, message queues, and other infrastructure clients. For those systems, forwarding the original request is tightly coupled to credentials, connection pools, client libraries, message acknowledgement semantics, retry policy, serialization, and framework-specific context. If MockServer or Mock SDK owns concrete forwarding, every new protocol will pull more runtime transport logic into a generic mock component.

The desired boundary is narrower: MockServer evaluates published rules and namespace fallback policy, Mock SDK exposes a typed decision, and injected code from `mockinject` decides how to execute that decision in the real technology stack.

HTTP is the only exception currently worth preserving. Direct HTTP runtime URLs can keep server-side original-request forwarding as a runtime debugging convenience for users who are testing rules and namespace fallback behavior. That HTTP forwarding path must be documented as runtime-only and must not become the Mock SDK or cross-protocol forwarding model.

## Solution

Reframe forwarding as a decision contract rather than an execution responsibility.

From the user's perspective:

- `mocksdk` can still normalize an HTTP request, call MockServer, and return either a response decision or a forward decision.
- `mocksdk` must not provide a helper that performs the original upstream request.
- `mockserver` must keep the SDK decision endpoint decision-only. It must never forward upstream traffic through that endpoint.
- direct HTTP runtime mock URLs may continue to execute forward fallback so users can debug HTTP rules without writing mockinject code.
- non-HTTP protocol integrations should use the same decision semantics but leave forwarding to protocol-specific mockinject code.

This keeps the generic tool chain small:

- MockServer owns matching, fallback policy, observability, and HTTP runtime debug behavior.
- Mock SDK owns normalization, configuration, endpoint calls, typed decisions, and response application helpers.
- Mockinject owns protocol-specific execution, including original-request forwarding.

## User Stories

1. As a mockinject author, I want Mock SDK to return a forward decision without executing it, so that injected code remains in control of protocol-specific forwarding.
2. As a mockinject author, I want the SDK package to avoid upstream forwarding helpers, so that it does not become coupled to one protocol client's transport model.
3. As a mockinject author, I want to inspect `DecisionKindForward` and decide what my injected adapter should do, so that HTTP, RPC, cache, and MQ integrations can each use their own stack.
4. As a mockinject author, I want request body replay to remain available after SDK normalization, so that injected HTTP code can still forward using its own implementation when appropriate.
5. As a mockinject author, I want response decisions to remain easy to apply for HTTP, so that response-mode mocks still require minimal glue.
6. As a MockServer user, I want the SDK decision endpoint to never forward upstream traffic, so that asking for a decision has no hidden side effects.
7. As a MockServer user, I want a forward decision to include enough metadata for injected code, so that mockinject can log fallback reason and honor timeout policy if it chooses.
8. As a MockServer user, I want direct HTTP runtime URLs to keep pass-through forwarding for local rule debugging, so that I can test namespace fallback behavior without building mockinject first.
9. As a MockServer user, I want runtime HTTP forwarding to be clearly labeled as debug/runtime behavior, so that I do not treat it as the general forwarding architecture.
10. As a MockServer user, I want future non-HTTP protocols to get decisions rather than MockServer-owned forwarding, so that credentials and client behavior stay inside the correct protocol adapter.
11. As a MockServer maintainer, I want concrete upstream forwarding removed from `mocksdk`, so that the public SDK surface stays stable and small.
12. As a MockServer maintainer, I want backend tests to prove SDK decision calls are side-effect free, so that future refactors do not accidentally add forwarding to the decision endpoint.
13. As a MockServer maintainer, I want backend code to make the HTTP runtime forwarding exception explicit, so that future protocol work does not reuse it accidentally.
14. As a MockServer maintainer, I want docs to explain the roles of MockServer, Mock SDK, and mockinject, so that contributors add protocol support at the right layer.
15. As a MockServer maintainer, I want all removed helper examples updated, so that users do not copy a deprecated forwarding path.

## Implementation Decisions

- Keep `DecisionKindForward` as a stable decision kind. The decision remains the cross-protocol contract.
- Keep the `ForwardDecision` payload for metadata such as timeout policy. It is policy information, not execution logic.
- Remove the Mock SDK helper that performs original HTTP request forwarding.
- Keep SDK request normalization and body restoration because they support decision generation and allow injected code to continue using the original request.
- Keep SDK response application helper because returning a mock response is still a first-class SDK responsibility for HTTP adapters.
- Keep the SDK decision endpoint decision-only. It must continue to evaluate matching and fallback without opening upstream connections.
- Preserve HTTP runtime forward fallback as a temporary and explicit runtime debugging convenience.
- Make backend naming or checks clear enough that runtime HTTP forwarding is not presented as the generic fallback execution model for future protocols.
- Do not remove namespace fallback type `forward`; it remains how users express pass-through policy.
- Do not introduce new protocol-specific forwarding implementations.
- Update documentation and examples so `mockinject` is responsible for protocol-specific forwarding.

## Testing Decisions

- Good tests should validate public behavior: SDK exported helpers, SDK decisions, endpoint side effects, and HTTP runtime debug forwarding.
- Mock SDK tests should prove response decision application still works and the deleted forwarding helper is no longer part of the tested public helper behavior.
- Mock SDK normalization tests should keep proving body replay after normalization because injected adapters depend on it.
- Backend controller tests should keep proving the SDK decision endpoint returns a forward decision without forwarding upstream traffic.
- Backend controller tests should keep proving direct HTTP runtime URLs can execute forward fallback for debugging.
- Backend service tests or controller tests should make the HTTP-only runtime forwarding boundary explicit where practical.
- Documentation changes should be verified by searching for removed helper names and outdated examples.
- The final implementation should pass `go test ./...`.

## Out of Scope

- Implementing mockinject.
- Implementing SPEX, gRPC, cache, or MQ protocol adapters.
- Removing the namespace fallback type `forward`.
- Removing typed forward decisions from the SDK or backend endpoint.
- Removing HTTP runtime debug forwarding.
- Adding credentials, connection pools, retries, circuit breakers, or protocol-specific forwarding behavior to MockServer.
- Adding a frontend workflow for protocol-specific injectors.

## Further Notes

This PRD intentionally narrows the previous Mock SDK package scope. The previous helper forwarding path was useful for proving HTTP request fidelity, but it is now too protocol-specific for the generic SDK boundary.

## Comments

- Implemented all three decision-only forward boundary issues.
- Mock SDK now returns forward decisions without exposing concrete upstream forwarding helpers.
- Backend forward fallback execution is explicitly scoped to HTTP runtime behavior; decision calls remain side-effect free.
- Docs now state that mockinject owns protocol-specific forwarding.
- Verified with `go test ./mocksdk`, `go test ./internal/...`, and `go test ./...`.
