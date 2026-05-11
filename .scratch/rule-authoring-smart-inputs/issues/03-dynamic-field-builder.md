# Dynamic field builder

Status: completed
Type: AFK

## Parent

.scratch/rule-authoring-smart-inputs/PRD.md

## What to build

Make dynamic protocol fields easier to author by letting users select a dynamic root and fill a focused key/path input. The resulting field path should still be the existing MockServer field path contract.

## Acceptance criteria

- [x] Dynamic HTTP query fields can be built from a query key.
- [x] Dynamic HTTP header fields can be built from a header key and normalize header keys to lower case.
- [x] Query and header fields support first-value and any-value modes.
- [x] Dynamic HTTP body fields can be built from a JSON body path.
- [x] Dynamic cache value fields can be built from a JSON value path.
- [x] The generated field path is visible in the predicate payload preview.

## Blocked by

- .scratch/rule-authoring-smart-inputs/issues/01-value-input-spec.md
