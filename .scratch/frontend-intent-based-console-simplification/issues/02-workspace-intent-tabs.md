# Workspace intent tabs

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-intent-based-console-simplification/PRD.md

## What to build

Group the existing detailed workbench modes under three user intents: Rules, Test, and Release. Keep existing deep links and detailed mode behavior, but present the primary task switch as intent-based instead of six parallel task tabs.

## Acceptance criteria

- [x] Workspace primary task switch shows Rules, Test, and Release.
- [x] Inspect and editor modes are grouped under Rules.
- [x] Simulation and result modes are grouped under Test.
- [x] Readiness and snapshots modes are grouped under Release.
- [x] Existing route query modes still open the correct detailed panel.
- [x] Intent grouping helper behavior is covered by unit tests.

## Blocked by

- .scratch/frontend-intent-based-console-simplification/issues/01-global-and-rulesets-noise-reduction.md

## Completion notes

Implemented in `workbenchTasks.ts`, `RuleSetRules.vue`, and helper tests. The primary workbench switch is now intent-based, while detailed modes remain addressable through existing query values and secondary controls.
