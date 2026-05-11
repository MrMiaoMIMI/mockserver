# Protocol spec catalog and rule validation

Status: completed
Type: AFK

## Parent

.scratch/multi-protocol-event-architecture/PRD.md

## What to build

Introduce code-registered protocol specs for HTTP and cache, expose them through the management API, and validate rule field paths and operators against those specs. A completed slice should let MockServer explain which fields and selectors each protocol supports while rejecting invalid rule fields before publish.

## Acceptance criteria

- [x] HTTP and cache protocol specs are code-registered.
- [x] Field specs include path, type, dynamic-path flag, and optional operators only.
- [x] Selector specs exist for HTTP and cache.
- [x] Operator defaults are derived from field type when a field does not override operators.
- [x] Rule validation rejects unknown field roots and invalid operators for known fields.
- [x] The management API exposes the registered protocol catalog.
- [x] Focused Go tests cover protocol catalog and validation behavior.

## Blocked by

None - can start immediately.

## Comments
