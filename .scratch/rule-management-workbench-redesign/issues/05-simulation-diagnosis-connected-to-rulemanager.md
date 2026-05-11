# Simulation Diagnosis Connected To RuleManager

Status: done
Type: AFK

## Parent

.scratch/rule-management-workbench-redesign/PRD.md

## What to build

Make simulation a contextual rule diagnosis workflow. Simulation should derive useful defaults from the current ruleset and selected rule, show structured diagnostic results, and feed match or miss signals back into RuleManager.

The completed slice should let users compare draft and published behavior without losing the current event input. Raw JSON should remain available, but structured diagnosis should be the default reading path.

## Acceptance criteria

- [x] Simulation default events are generated from current namespace, protocol, selector values, and selected rule hints when safe.
- [x] Users can run simulation against draft and published behavior while preserving the current event input.
- [x] Simulation results show selector match, rule match, fallback state, response result, candidate rules, and miss reason as structured diagnostics.
- [x] RuleManager reflects the latest simulation status for matched, missed, skipped, selector miss, rule miss, or fallback states where the response data supports it.
- [x] Raw simulation request and response JSON remain inspectable after structured summaries.
- [x] Simulation failures are shown in the simulation workflow instead of as generic page errors.
- [x] Tests cover simulation default event generation and diagnostic view-model mapping for match, selector miss, rule miss, fallback, and API failure.
- [x] Frontend build/type verification passes.

## Blocked by

- .scratch/rule-management-workbench-redesign/issues/02-rulemanager-browse-search-filter-selection.md
- .scratch/rule-management-workbench-redesign/issues/04-selected-rule-workbench-overview-editor.md

## Comments

- Completed in sequence as part of the Rule Management Workbench redesign implementation. Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for desktop and narrow layouts.
