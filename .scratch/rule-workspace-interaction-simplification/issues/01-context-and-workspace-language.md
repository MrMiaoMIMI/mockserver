# Context actions and workspace language cleanup

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/rule-workspace-interaction-simplification/PRD.md

## What to build

Remove redundant ruleset context actions and rename the right-side focused work area so the page reads as a clear Rule Workspace. Back-to-list and refresh behavior should remain available through existing shell/browser controls, while ruleset details stay available through a direct lightweight toggle.

## Acceptance criteria

- [x] The ruleset context area no longer exposes Back to list.
- [x] The ruleset context area no longer exposes Refresh data.
- [x] The page still supports navigation back to Rulesets through the sidebar or browser back.
- [x] The page still supports data refresh through the global refresh control.
- [x] Ruleset details remain available through a direct details toggle.
- [x] The right-side area is labeled Rule Workspace instead of active task.
- [x] Current workflow title and context text remain visible.

## Blocked by

None - can start immediately.

## Completion notes

Implemented on the Ruleset Workspace page. The normal context action area now shows Settings and a direct Details toggle, while shell navigation/refresh remain the canonical paths. The right-side dock now uses Rule Workspace language.
