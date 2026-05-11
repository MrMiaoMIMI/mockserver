# Compact rule cards

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-intent-based-console-simplification/PRD.md

## What to build

Make Rule Manager cards easier to scan by demoting priority and condition details. Rule cards should default to identity, state, diagnostics, and action summary, with condition details available through a simple disclosure.

## Acceptance criteria

- [x] Priority no longer occupies a large block in every rule card.
- [x] Condition details are hidden by default and can be expanded per row.
- [x] Rule identity, enabled state, diagnostic state, and action summary remain visible.
- [x] Existing rule selection, edit, priority, enable/disable, duplicate, delete, and move flows still work.
- [x] Row controls remain accessible and do not trigger unwanted row selection.

## Blocked by

- .scratch/frontend-intent-based-console-simplification/issues/02-workspace-intent-tabs.md

## Completion notes

Implemented in `RuleManager.vue`. Rule cards now default to identity, state, diagnostics, and action summary. Priority is a compact label, and condition detail is available through a disclosure control on each card.
