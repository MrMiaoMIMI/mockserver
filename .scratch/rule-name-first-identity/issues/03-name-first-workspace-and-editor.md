# Name-first workspace and editor

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/rule-name-first-identity/PRD.md

## What to build

Update Rule Workspace titles, inspect panels, editor identity copy, duplicate behavior, and create/edit form messaging so Rule Name is the user-facing identity and Rule ID remains stable metadata.

## Acceptance criteria

- [x] Workbench titles use Rule Name first for inspect and edit.
- [x] Inspect panel uses Rule Name as the selected-rule title and Rule ID as metadata.
- [x] Create form requires Name and shows generated Rule ID as secondary metadata.
- [x] Edit form keeps Rule ID read-only and lets Name be edited.
- [x] Duplicate creates a copied Rule Name and unique generated Rule ID.
- [x] Editor readiness and error copy refer to Rule Name appropriately.

## Completion notes

- Updated workspace titles, active context labels, and inspect panel to use Rule Name first.
- Updated create/edit identity UI so create is name-first with generated ID preview and edit keeps ID read-only.
- Updated duplicate behavior to copy the display name and generate a unique technical ID from that copied name.

## Blocked by

- .scratch/rule-name-first-identity/issues/01-required-rule-name-contract.md
- .scratch/rule-name-first-identity/issues/02-name-first-rule-manager-display.md
