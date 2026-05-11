# SDK protocol adapters

Status: completed
Type: AFK

## Parent

.scratch/protocol-end-to-end-adaptation/PRD.md

## What to build

Make the Mock SDK's core/adapter separation explicit. The SDK core should remain a generic decision client while HTTP and cache adapter helpers own protocol-specific event construction and response application.

## Acceptance criteria

- [x] SDK core types and client remain protocol-neutral.
- [x] HTTP adapter helpers own HTTP request normalization and HTTP response application.
- [x] Cache adapter helpers own cache event construction.
- [x] Existing top-level helper compatibility is removed or reduced according to the greenfield no-compatibility rule.
- [x] SDK tests prove HTTP and cache adapters can create events and call the decision client.
- [x] Documentation explains where future SPEX/gRPC/MQ adapters should live.

## Blocked by

- .scratch/protocol-end-to-end-adaptation/issues/01-generic-selector-contract.md

