# Mock SDK forward decision without forwarding helper

Status: done
Type: AFK

## Parent

.scratch/decision-only-forward-boundary/PRD.md

## What to build

Remove concrete original-request forwarding from Mock SDK while preserving the typed decision workflow. Mock SDK should still return `forward` decisions and preserve normalized HTTP request data, but it should no longer expose a helper that performs the upstream request for the caller.

Covers user stories: 1, 2, 3, 4, 5, 11.

## Acceptance criteria

- [x] Mock SDK no longer exposes or tests a helper that forwards the original upstream HTTP request.
- [x] `DecisionKindForward` and `ForwardDecision` remain available as typed decision contracts.
- [x] `Decide` and `DecideHTTP` still return forward decisions from the MockServer decision endpoint.
- [x] HTTP request normalization still restores the original request body for injected code.
- [x] The response decision helper still writes status, headers, and body for response-mode mocks.
- [x] `go test ./mocksdk` passes.

## Blocked by

None - can start immediately.

## Comments

- Removed the SDK `ForwardOriginal` helper and its forwarding test.
- Kept `DecisionKindForward`, `ForwardDecision`, `Decide`, `DecideHTTP`, request body replay, and `ApplyResponse`.
- Verified with `go test ./mocksdk`.
