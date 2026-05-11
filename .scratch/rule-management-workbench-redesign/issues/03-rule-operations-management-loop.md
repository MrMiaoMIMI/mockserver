# Rule Operations Management Loop

Status: done
Type: AFK

## Parent

.scratch/rule-management-workbench-redesign/PRD.md

## What to build

Complete the end-to-end rule management loop from RuleManager. Users should be able to create, duplicate, enable, disable, reorder, and delete rules from the rule management surface, with clear target labels and predictable state updates.

This slice should keep existing admin API contracts and immediate-save semantics unless implementation evidence proves a narrow backend improvement is required. The completed slice should make repeated rule operations efficient and safe for a ruleset with many rules.

## Acceptance criteria

- [x] Users can create a new rule from RuleManager and land in the correct Workbench task for completing it.
- [x] Users can duplicate an existing rule with a generated non-conflicting rule ID and clear selected-rule handoff.
- [x] Users can enable and disable rules with visible state updates in RuleManager and Workbench.
- [x] Users can reorder rules or update priority with visible priority consequences.
- [x] Users can delete a rule only after confirming the exact target name or ID.
- [x] Operation loading and failure states are scoped to the relevant rule action where possible.
- [x] Rule operation behavior preserves existing backend API contracts and rule payload shape.
- [x] Tests cover duplicate ID generation, operation state derivation, and stable selected-rule behavior after mutation.
- [x] Frontend build/type verification passes.

## Blocked by

- .scratch/rule-management-workbench-redesign/issues/02-rulemanager-browse-search-filter-selection.md

## Comments

- Completed in sequence as part of the Rule Management Workbench redesign implementation. Verified with `npm run test`, `npm run type-check`, `npm run build`, `go test ./...`, and Playwright browser checks for desktop and narrow layouts.
