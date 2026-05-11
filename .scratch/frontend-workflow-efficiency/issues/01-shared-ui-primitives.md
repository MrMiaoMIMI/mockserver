# Shared UI primitives

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-workflow-efficiency/PRD.md

## What to build

Create reusable UI primitives for repeated frontend patterns: metrics, state chips, segmented filters, key/value details, task switching, and empty/action panels. The completed slice should immediately reduce duplication in at least one existing page while keeping each primitive presentation-focused and simple to reuse.

## Acceptance criteria

- [x] Shared primitives exist for metric cards, state chips, segmented filters, key/value grids, and task rail style switching.
- [x] The primitives use the existing design tokens and do not introduce a second visual system.
- [x] The primitives expose small stable props and events instead of page-specific business logic.
- [x] At least one existing page uses the new primitives end to end.
- [x] Frontend build passes.

## Completion notes

- Added `MetricCard`, `StateChip`, `FilterSegment`, `KeyValueGrid`, and `TaskRail`.
- Reused the primitives across Rulesets, Rule Manager, Ruleset Workspace, and Runtime Diagnostics.
- Verified through `npm run build`.

## Blocked by

None - can start immediately.
