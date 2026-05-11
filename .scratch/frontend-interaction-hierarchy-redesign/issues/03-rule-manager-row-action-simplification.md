# Rule manager row action simplification

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-interaction-hierarchy-redesign/PRD.md

## What to build

Reduce visual noise in Rule Manager cards by keeping frequent actions visible and moving less common or destructive actions into a row menu. Preserve the full rule management loop while making rule rows easier to scan.

## Acceptance criteria

- [x] Rule cards expose the primary rule action clearly.
- [x] Move priority, duplicate, delete, and other secondary operations are available through a compact menu or equivalent secondary affordance.
- [x] Enabled/disabled and diagnostic state remain readable without crowding the row.
- [x] Existing rule selection, edit, priority changes, enable/disable, duplicate, and delete flows still work.
- [x] Row controls have accessible labels and do not trigger row selection unexpectedly.

## Completion notes

- Kept Edit visible as the primary row operation.
- Moved move up/down, set priority, enable/disable, duplicate, and delete into a row More menu.
- Preserved accessible labels and stop-propagation behavior for row controls.

## Blocked by

- .scratch/frontend-interaction-hierarchy-redesign/issues/02-workspace-task-first-controls.md
