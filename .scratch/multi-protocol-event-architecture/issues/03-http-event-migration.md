# HTTP event migration

Status: completed
Type: AFK

## Parent

.scratch/multi-protocol-event-architecture/PRD.md

## What to build

Migrate HTTP normalization, runtime handling, SDK helpers, docs, and frontend types to the new generic event request shape. HTTP JSON bodies should be available at `request.body`, raw bodies at `request.raw_body`, and existing HTTP rules should use the new event document consistently.

## Acceptance criteria

- [x] MockServer HTTP runtime adapter creates the new request document.
- [x] Mock SDK HTTP normalizer creates the new request document.
- [x] JSON body sets both parsed body and raw body; non-JSON body sets raw body only; empty body sets neither.
- [x] Existing HTTP controller, engine, and SDK tests pass after migration.
- [x] Documentation examples use `request.body` and `request.raw_body`.
- [x] Frontend TypeScript types/build accept the generic event request shape.

## Blocked by

- .scratch/multi-protocol-event-architecture/issues/02-generic-event-request-document.md

## Comments
