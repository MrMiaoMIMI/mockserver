# Value input spec engine

Status: completed
Type: AFK

## Parent

.scratch/rule-authoring-smart-inputs/PRD.md

## What to build

Create the protocol-aware smart value input model that converts a selected field and operator into an input specification. This spec should define the editor kind, whether a value is required, helper text, placeholder, examples, and normalization behavior so rule and selector authoring can share one contract.

## Acceptance criteria

- [x] Field type plus operator maps to the expected editor kind for text, number, boolean, JSON, list, regex, and no-value checks.
- [x] `exists`, `not_exists`, `is_null`, and `is_not_null` mark value as not required.
- [x] `in` and `not_in` mark the value editor as list and serialize to arrays.
- [x] Numeric comparison operators produce numeric editor guidance.
- [x] HTTP and cache field paths produce useful placeholders and examples.
- [x] Unit tests cover HTTP method/path/header/body and cache operation/key/value examples.

## Blocked by

None - can start immediately.
