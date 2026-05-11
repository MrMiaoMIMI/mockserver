# Responsive, Accessibility, And Browser QA Hardening

Status: done
Type: AFK

## Parent

.scratch/rule-management-workbench-redesign/PRD.md

## What to build

Harden the redesigned rule management experience across viewport sizes, long technical values, keyboard interaction, focus behavior, loading states, empty states, and realistic browser QA scenarios. This slice should make the completed redesign reliable enough for daily rule management.

Use Playwright or equivalent browser automation to verify the actual page behavior. The focus is not pixel-perfect theme testing; the focus is whether the information architecture and workflows remain usable under realistic conditions.

## Acceptance criteria

- [x] Desktop layout gives primary space to RuleManager and contextual Workbench.
- [x] Narrow viewport layout preserves the task order: compact ruleset context, primary actions, RuleManager, active Workbench.
- [x] Long rule IDs, selector values, paths, headers, and response snippets are inspectable without overlapping or breaking layout.
- [x] Keyboard focus order is usable across RuleManager rows, filters, operations, Workbench modes, and confirmations.
- [x] Visible focus states exist for custom interactive controls.
- [x] Browser QA covers no rules, one rule, many rules, disabled rule, invalid rule, long technical values, create, duplicate, edit, reorder, delete, simulate match, selector miss, rule miss, fallback, validate, publish, snapshot preview, and rollback preview.
- [x] Regression QA confirms namespace fallback behavior and publish or rollback behavior are unchanged.
- [x] Production build/type verification passes.
- [x] Any remaining UX risks are documented in issue comments with screenshots or exact reproduction notes.

## Blocked by

- .scratch/rule-management-workbench-redesign/issues/03-rule-operations-management-loop.md
- .scratch/rule-management-workbench-redesign/issues/04-selected-rule-workbench-overview-editor.md
- .scratch/rule-management-workbench-redesign/issues/05-simulation-diagnosis-connected-to-rulemanager.md
- .scratch/rule-management-workbench-redesign/issues/06-validation-publish-readiness-workflow.md
- .scratch/rule-management-workbench-redesign/issues/07-snapshot-rollback-contextual-workflow.md

## Comments

- Completed in sequence as part of the Rule Management Workbench redesign implementation. Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for desktop and narrow layouts.
