# Ruleset workspace task flow

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-workflow-efficiency/PRD.md

## What to build

Refine the Ruleset Workspace so it reads as a task flow instead of unrelated panels. The user should understand the selected rule context, the available workbench tasks, and the next action path from inspect to edit, simulate, validate, publish, snapshots, and result inspection.

## Acceptance criteria

- [x] Workbench mode switching uses a clearer task rail presentation.
- [x] Selected rule context remains visible while using editor, simulation, readiness, snapshot, and result modes.
- [x] Rule Manager filters and status counts use shared primitives where appropriate.
- [x] Existing rule operations, simulation, readiness, publish, snapshots, rollback, and result panels remain functional.
- [x] Desktop workspace browser QA shows clearer hierarchy and no layout overlap.
- [x] Narrow viewport behavior remains usable.

## Completion notes

- Replaced the local workbench switch with `TaskRail` and kept selected-rule context in the tool dock header.
- Replaced Rule Manager local segmented controls/status badges with shared primitives.
- Browser QA covered the Ruleset Workspace and task rail hierarchy.

## Blocked by

- .scratch/frontend-workflow-efficiency/issues/01-shared-ui-primitives.md
