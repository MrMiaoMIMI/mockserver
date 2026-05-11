# Snapshot And Rollback Contextual Workflow

Status: done
Type: AFK

## Parent

.scratch/rule-management-workbench-redesign/PRD.md

## What to build

Make snapshot history and rollback preview a contextual Workbench workflow that supports rule management without dominating the default page. Users should be able to inspect published history, preview rollback impact, and execute rollback with clear target context.

This slice should preserve existing snapshot and rollback semantics. It should improve the interaction structure and result presentation rather than changing storage or publish history contracts.

## Acceptance criteria

- [x] Snapshot history is available from the Workbench without taking over the default RuleManager view.
- [x] Snapshot rows show high-signal metadata such as snapshot ID, version, published time, operator, reason, and source snapshot when available.
- [x] Selecting a snapshot updates rollback preview context clearly.
- [x] Rollback preview distinguishes current state from target snapshot state before rollback.
- [x] Rollback confirmation names the exact target snapshot or version.
- [x] Rollback result is shown as a structured outcome before raw JSON.
- [x] Existing snapshot, rollback preview, and rollback API contracts are preserved.
- [x] Tests cover snapshot history view-model mapping and rollback preview/result mapping.
- [x] Frontend build/type verification passes.

## Blocked by

- .scratch/rule-management-workbench-redesign/issues/01-compact-ruleset-context-workspace-task-model.md
- .scratch/rule-management-workbench-redesign/issues/02-rulemanager-browse-search-filter-selection.md

## Comments

- Completed in sequence as part of the Rule Management Workbench redesign implementation. Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for desktop and narrow layouts.
