# Rule stack selection and readable rule summaries

Status: done
Type: AFK

## Parent

.scratch/ruleset-workspace-refactor/PRD.md

## What to build

Refactor the rule stack into the primary authoring surface for a ruleset. Users should be able to scan rules, select a rule, understand condition and action behavior from readable summaries, and perform common rule operations without opening every rule editor.

This slice should keep rule mutations immediate-save through existing admin endpoints. It should add or extract summary helpers so condition and action language is consistent across rule rows and later detail/diagnostic panels.

## Acceptance criteria

- [x] The rule stack supports selecting an active rule and visibly distinguishes the selected rule.
- [x] Each rule row shows priority, enabled state, rule ID, name, condition summary, action summary, and stable operation controls.
- [x] Condition summaries handle predicate, ALL, ANY, NOT, and CEL expression shapes.
- [x] Action summaries handle static response, template response, CEL response, sequence response, and webhook response shapes.
- [x] Users can enable, disable, update priority, duplicate, and delete rules from the rule stack with clear target labels and confirmation for destructive actions.
- [x] Rule stack layout handles long rule IDs and long summaries without layout jumps.
- [x] Frontend build/type checks pass.

## Blocked by

- .scratch/ruleset-workspace-refactor/issues/01-ruleset-workspace-context-baseline.md

## Comments

- Completed with immediate-save mutations through existing admin APIs.
- Verified with `npm run build`, `git diff --check`, and Chrome headless desktop/narrow screenshots for the target ruleset workspace.
