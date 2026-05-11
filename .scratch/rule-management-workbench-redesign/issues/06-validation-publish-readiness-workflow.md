# Validation And Publish Readiness Workflow

Status: done
Type: AFK

## Parent

.scratch/rule-management-workbench-redesign/PRD.md

## What to build

Turn validation and publishing into a clear Workbench workflow. Users should understand whether the draft is valid, whether it differs from the published version, what blocks publishing, and what happened after a publish action.

This slice should connect validation results to the selected rule and RuleManager where possible, while keeping raw backend responses available for deep debugging. Existing publish and validation API contracts should be preserved.

## Acceptance criteria

- [x] Workbench has a validation and publish readiness workflow that is distinct from simulation and snapshot tasks.
- [x] Users can run validation and see structured success, warning, or failure output.
- [x] Validation output links back to relevant rule, condition, action, or raw JSON sections where response data supports it.
- [x] RuleManager can show validation status signals after validation runs.
- [x] Publish readiness summarizes validation state, draft-vs-published state, and publish availability.
- [x] Publish result is shown as a structured outcome before raw JSON.
- [x] Validation and publish failures are scoped to the workflow that produced them.
- [x] Tests cover validation and publish diagnostic view-model mapping.
- [x] Frontend build/type verification passes.

## Blocked by

- .scratch/rule-management-workbench-redesign/issues/02-rulemanager-browse-search-filter-selection.md
- .scratch/rule-management-workbench-redesign/issues/04-selected-rule-workbench-overview-editor.md

## Comments

- Completed in sequence as part of the Rule Management Workbench redesign implementation. Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for desktop and narrow layouts.
