# Generic event request document

Status: completed
Type: AFK

## Parent

.scratch/multi-protocol-event-architecture/PRD.md

## What to build

Migrate MockServer's Event request model from a fixed HTTP request struct to a generic request document. Matching, CEL input, templates, and field resolution should evaluate `request.*` paths against the document while preserving common event metadata.

## Acceptance criteria

- [x] Event request is represented as a dynamic request document.
- [x] Existing rule matching resolves fields from the generic request document.
- [x] `exists`, `not_exists`, `is_null`, and `is_not_null` have the agreed semantics.
- [x] Dynamic JSON subfields can use the full first-version operator set.
- [x] Existing HTTP behavior is migrated directly without old-shape compatibility adapters.
- [x] Go tests cover field resolution and null/missing semantics.

## Blocked by

- .scratch/multi-protocol-event-architecture/issues/01-protocol-spec-catalog-and-rule-validation.md

## Comments
