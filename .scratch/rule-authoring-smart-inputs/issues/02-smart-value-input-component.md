# Smart value input component

Status: completed
Type: AFK

## Parent

.scratch/rule-authoring-smart-inputs/PRD.md

## What to build

Build a reusable smart value input component that renders the correct input control from a value input specification and emits correctly typed values. It should also show helper text, examples, and lightweight validation feedback.

## Acceptance criteria

- [x] The component supports no-value, text, number, boolean, JSON, list, and regex editors.
- [x] List input allows adding and removing values without hand-writing JSON arrays.
- [x] Regex input shows validation feedback for invalid patterns.
- [x] JSON input preserves typed object/array/primitive values when valid.
- [x] Example values can be applied with one click.
- [x] The component stays compact enough for condition and selector rows.

## Blocked by

- .scratch/rule-authoring-smart-inputs/issues/01-value-input-spec.md
