# Rule card action reduction

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/rule-workspace-interaction-simplification/PRD.md

## What to build

Reduce rule-card action noise by keeping card actions for structural operations and moving property edits into Edit. The card menu should focus on order, duplicate, and delete.

## Acceptance criteria

- [x] Rule card primary action remains Edit.
- [x] Rule card More menu keeps Move up, Move down, Duplicate, and Delete.
- [x] Rule card More menu no longer shows Set priority.
- [x] Rule card More menu no longer shows Enable or Disable.
- [x] Priority can still be changed in the editor.
- [x] Enabled state can still be changed in the editor.
- [x] Existing duplicate, delete, and move behavior still works.

## Blocked by

- .scratch/rule-workspace-interaction-simplification/issues/02-compact-rule-manager-filters.md

## Completion notes

Implemented in Rule Manager. Card actions are now focused on Edit plus structural More actions: move up, move down, duplicate, and delete. Priority and enabled state remain in the editor identity section.
