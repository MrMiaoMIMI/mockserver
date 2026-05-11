# JSON readability upgrade

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/rule-workspace-interaction-simplification/PRD.md

## What to build

Upgrade the shared JSON editing and viewing experience so request, result, rule, condition, and preview JSON are readable in the PC console. The shared JSON surface should provide high-contrast code styling, line numbers, copy, format, and expanded viewing.

## Acceptance criteria

- [x] Shared JSON editor has higher contrast and clearer code typography.
- [x] JSON editor displays line numbers.
- [x] JSON editor exposes copy and format actions.
- [x] JSON editor supports an expanded viewing/editing mode.
- [x] Editable JSON remains editable.
- [x] Read-only raw JSON remains readable through the same shared component.
- [x] Test tab request JSON is visibly improved.
- [x] Result raw JSON is visibly improved.

## Blocked by

- .scratch/rule-workspace-interaction-simplification/issues/04-name-first-rule-creation.md

## Completion notes

Implemented in the shared JSON editor. JSON surfaces now use a high-contrast editor body, synced line-number gutter, Format and Copy actions, and an Expand dialog that works for editable and read-only JSON usage.
