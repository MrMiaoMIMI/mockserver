# RuleManager Browse, Search, Filter, And Selection

Status: done
Type: AFK

## Parent

.scratch/rule-management-workbench-redesign/PRD.md

## What to build

Turn RuleManager into the primary rule browsing and selection surface. Users should be able to scan, search, filter, and select rules without losing context, and the selected rule should drive the Workbench consistently.

This slice should introduce or strengthen rule collection view helpers for visible rows, search matching, filters, stable selection, empty states, status counts, and readable summaries. It should keep the underlying priority order intact and only change the visible presentation when search or filters are active.

## Acceptance criteria

- [x] RuleManager gives the rule list enough visual priority to feel like the main management surface.
- [x] Each rule row shows priority, enabled state, name, rule ID, condition summary, action summary, and high-signal status indicators.
- [x] Users can search rules by name, ID, condition summary, action summary, and status text.
- [x] Users can filter rules by enabled state and action type, with room for validation or simulation status filters when those signals exist.
- [x] Filtering and search do not mutate rule priority order.
- [x] Selected rule state remains stable after filtering, search changes, refresh, and row updates where possible.
- [x] Long rule IDs and technical values remain inspectable without breaking row layout.
- [x] Rule collection view helpers have focused tests for search, filters, selected-row stability, ordering, and empty states.
- [x] Frontend build/type verification passes.

## Blocked by

- .scratch/rule-management-workbench-redesign/issues/01-compact-ruleset-context-workspace-task-model.md

## Comments

- Completed in sequence as part of the Rule Management Workbench redesign implementation. Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for desktop and narrow layouts.
