# Selector authoring smart inputs

Status: completed
Type: AFK

## Parent

.scratch/rule-authoring-smart-inputs/PRD.md

## What to build

Apply the same smart value input behavior to ruleset selector authoring so selector configuration and rule condition configuration behave consistently.

## Acceptance criteria

- [x] Selector value editor changes based on selector field and operator.
- [x] No-value selector operators omit `value` from the generated selector condition.
- [x] Selector rows show helper text, examples, and payload preview.
- [x] Selector validation uses typed smart input values instead of reparsing raw strings.
- [x] HTTP host/path and cache operation/key selector examples are easy to fill.
- [x] Existing ruleset settings save flow still produces valid `selector.all` payloads.

## Blocked by

- .scratch/rule-authoring-smart-inputs/issues/02-smart-value-input-component.md
