# Compact Rule Manager filters

Status: done
Type: AFK
Completed: 2026-05-11

## Parent

.scratch/rule-workspace-interaction-simplification/PRD.md

## What to build

Compress the Rule Manager filter area so the rule stack receives more vertical space. Search should remain inline, while enabled/action/diagnostic filters move into a popover with visible active-filter chips and a reset path.

## Acceptance criteria

- [x] Rule Manager shows one compact toolbar row by default.
- [x] Search remains directly visible.
- [x] Enabled, action type, and diagnostic filters are available through a Filters popover.
- [x] Non-default filters are summarized as chips.
- [x] Users can reset filters from the toolbar or popover.
- [x] Rule status counts are retained in a compact summary.
- [x] Existing filter behavior still narrows the visible rule stack correctly.

## Blocked by

- .scratch/rule-workspace-interaction-simplification/issues/01-context-and-workspace-language.md

## Completion notes

Implemented in Rule Manager. The toolbar is now search-first with a Filters popover for secondary filters, active filter chips, clear/reset behavior, and a compact shown/enabled/invalid/matched summary.
