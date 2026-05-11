# Selected Rule Workbench Overview And Editor

Status: done
Type: AFK

## Parent

.scratch/rule-management-workbench-redesign/PRD.md

## What to build

Redesign the selected-rule Workbench so it starts with a useful rule overview and opens into a spacious structured editor when needed. The Workbench should be driven by the selected rule and should avoid cramped editing surfaces or duplicate headers.

This slice should preserve raw JSON editing for advanced users while making common identity, condition, and action edits easier through structured sections. UI form changes must remain isolated behind rule payload adapters so backend rule contracts are not accidentally changed.

## Acceptance criteria

- [x] Selecting a rule opens or updates a selected-rule overview in the Workbench.
- [x] The overview summarizes identity, priority, enabled state, condition behavior, action behavior, and relevant status signals.
- [x] Editing uses a spacious Workbench mode rather than a cramped modal or narrow side panel.
- [x] Rule editing separates identity, matching condition, response action, and raw JSON or advanced editing.
- [x] Condition editing supports structured predicate, ALL, ANY, NOT, CEL expression, and raw JSON forms.
- [x] Action editing exposes relevant controls for static, template, CEL, sequence, and webhook responses.
- [x] Validation errors point to the relevant field or section before submitting to the backend.
- [x] Rule form adapter tests prove API rule payload shape is preserved for create and update.
- [x] Frontend build/type verification passes.

## Blocked by

- .scratch/rule-management-workbench-redesign/issues/01-compact-ruleset-context-workspace-task-model.md
- .scratch/rule-management-workbench-redesign/issues/02-rulemanager-browse-search-filter-selection.md

## Comments

- Completed in sequence as part of the Rule Management Workbench redesign implementation. Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for desktop and narrow layouts.
