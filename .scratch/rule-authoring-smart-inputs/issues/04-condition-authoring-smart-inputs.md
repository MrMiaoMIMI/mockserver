# Condition authoring smart inputs

Status: completed
Type: AFK

## Parent

.scratch/rule-authoring-smart-inputs/PRD.md

## What to build

Integrate smart inputs and dynamic field building into rule condition authoring. Users should be guided by protocol fields and operators, with typed value controls and payload previews.

## Acceptance criteria

- [x] Condition value editor changes when field or operator changes.
- [x] Operator changes clear or preserve values according to whether the operator needs a value.
- [x] Dynamic field builder is available for dynamic protocol fields.
- [x] Predicate payload preview shows the exact condition payload.
- [x] Existing HTTP presets continue to work as shortcuts.
- [x] Frontend tests cover condition payloads for HTTP and cache smart inputs.

## Blocked by

- .scratch/rule-authoring-smart-inputs/issues/02-smart-value-input-component.md
- .scratch/rule-authoring-smart-inputs/issues/03-dynamic-field-builder.md
