# Workspace task-first controls

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-interaction-hierarchy-redesign/PRD.md

## What to build

Simplify Ruleset Workspace around the active task and selected rule. Make the context bar more compact, convert the task rail into a lighter desktop toolbar, and move publish into the readiness context so users are guided to validate before releasing.

## Acceptance criteria

- [x] Workspace context bar shows the current ruleset identity and essential context without pushing the work area down.
- [x] Publish is not exposed as a constant top-level action outside the readiness workflow.
- [x] Task switching is compact, clear, keyboard reachable, and visually consistent with the rest of the console.
- [x] The selected rule and current task remain obvious while inspecting, editing, simulating, checking readiness, viewing snapshots, or viewing results.
- [x] Existing edit, simulate, readiness, publish, snapshot, rollback, and result workflows remain available.

## Completion notes

- Replaced the constant Publish header action with a Readiness task entry.
- Converted task switching from large cards into a compact segmented toolbar.
- Renamed the readiness task label and title to focus on validation before publish.

## Blocked by

- .scratch/frontend-interaction-hierarchy-redesign/issues/01-rulesets-entry-information-hierarchy.md
