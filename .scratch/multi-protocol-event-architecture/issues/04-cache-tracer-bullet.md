# Cache tracer bullet

Status: completed
Type: AFK

## Parent

.scratch/multi-protocol-event-architecture/PRD.md

## What to build

Add the first non-HTTP tracer bullet with cache event support. Users should be able to define a cache ruleset, match cache operations by operation/key/value fields, and receive a decision through the same SDK decision endpoint used by HTTP.

## Acceptance criteria

- [x] Mock SDK exposes a cache event builder.
- [x] Cache request fields include operation, key, ttl_ms, and value where applicable.
- [x] A cache ruleset can match a published cache event through the decision endpoint.
- [x] Cache rule miss can return a forward decision through namespace fallback.
- [x] Tests prove cache does not rely on HTTP host/path/method.
- [x] Documentation explains the cache tracer bullet and mockinject ownership.

## Blocked by

- .scratch/multi-protocol-event-architecture/issues/01-protocol-spec-catalog-and-rule-validation.md
- .scratch/multi-protocol-event-architecture/issues/02-generic-event-request-document.md

## Comments
