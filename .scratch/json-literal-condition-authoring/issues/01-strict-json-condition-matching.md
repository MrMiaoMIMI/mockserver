# Strict JSON condition matching

Status: completed
Type: AFK

## Parent

.scratch/json-literal-condition-authoring/PRD.md

## What to build

Make backend condition evaluation match the JSON literal contract. Equality and list membership should preserve JSON type semantics, and numeric comparisons should only succeed for numeric values.

## Acceptance criteria

- [x] `123` does not equal `"123"`.
- [x] `true` does not equal `"true"`.
- [x] JSON numbers compare equal across Go numeric representations.
- [x] Arrays and objects compare structurally for equality and membership.
- [x] Numeric comparison operators do not parse numeric-looking strings.
- [x] Tests cover strict equality, `in`, `not_in`, and numeric comparisons.

## Completion notes

- Implemented strict JSON comparison in `internal/engine/matcher.go`.
- Covered strict equality, `in`, `not_in`, arrays, objects, and numeric comparison behavior in `internal/engine/matcher_test.go`.

## Blocked by

None - can start immediately.
