# Compact Ruleset Context And Workspace Task Model

Status: done
Type: AFK

## Parent

.scratch/rule-management-workbench-redesign/PRD.md

## What to build

Reshape the ruleset rules page around a compact ruleset context and an explicit workspace task model. Users should immediately understand which ruleset they are editing, while most of the available screen space remains reserved for rule management and the contextual Workbench.

This slice should make RuleManager and Workbench responsibilities explicit. It should establish first-class state for the active rule, active Workbench task, ruleset context summary, loading states, operation errors, and compact metadata disclosure. The completed slice should be demoable even before later rule operations are improved.

## Acceptance criteria

- [x] The rules page shows ruleset identity, namespace, protocol, selector summary, draft version, publish state, and rule count in a compact context area.
- [x] Detailed ruleset metadata is available on demand without taking persistent vertical space.
- [x] The page clearly separates ruleset context, RuleManager, and Workbench responsibilities.
- [x] Active rule and active Workbench task are explicit state concepts.
- [x] Workbench modes can represent at least inspect, edit, simulate, validate, publish, snapshots, and rollback tasks.
- [x] Loading, empty, not-found, and API error states remain visible and understandable.
- [x] Existing route and backend API contracts are preserved.
- [x] Frontend build/type verification passes.

## Blocked by

None - can start immediately.

## Comments

- Completed in sequence as part of the Rule Management Workbench redesign implementation. Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for desktop and narrow layouts.
