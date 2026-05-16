# JSON literal value authoring

Status: completed
Type: AFK

## Parent

.scratch/json-literal-condition-authoring/PRD.md

## What to build

Update rule condition value authoring so dynamic JSON fields use a JSON literal editor. Users should type the value exactly as JSON, see validation errors for invalid literals, and see the resulting condition payload preview.

## Acceptance criteria

- [x] Dynamic JSON child fields use JSON literal parsing for operators that need a value.
- [x] `123`, `"123"`, `true`, `null`, arrays, and objects save as their parsed JSON values.
- [x] Unquoted plain strings show a direct invalid JSON error.
- [x] No-value operators still remove the value.
- [x] Existing non-dynamic text, number, boolean, regex, list, query, and header inputs remain usable.
- [x] Unit tests cover value input specs and rule payload conversion.

## Completion notes

- Dynamic JSON roots and descendants now use JSON literal editing in `valueInputSpec`.
- Invalid JSON literals are kept in editor state for actionable errors and blocked before payload build.

## Blocked by

- .scratch/json-literal-condition-authoring/issues/01-strict-json-condition-matching.md
