# Ruleset List Entry Redesign

Status: done
Type: AFK

## Parent

.scratch/ruleset-namespace-entry-experience/PRD.md

## What to build

Redesign Ruleset List as a decision-oriented entry surface. Users should be able to search, filter, sort, compare state, inspect selector summaries, and open the rules workspace as the primary action.

## Acceptance criteria

- [x] Ruleset List shows high-signal metrics for total, shown, published, changed, draft-only, enabled, and empty-rule states.
- [x] Users can search by name, ID, namespace, host, path, publish state, and rule IDs.
- [x] Users can filter by enabled state, publish state, namespace, and rule presence.
- [x] Users can sort by name, namespace, rule count, publish state, and version.
- [x] Each ruleset row/card shows name, ID, namespace, protocol, publish state, enabled state, rule count, selector hosts, selector paths, and version.
- [x] Opening the rules workspace is the primary action; settings and publish remain available as secondary actions.
- [x] Long IDs and selector values remain inspectable without breaking layout.
- [x] Frontend type-check, tests, and build pass.

## Blocked by

- .scratch/ruleset-namespace-entry-experience/issues/01-entry-view-models.md

## Comments

- Redesigned `web/src/views/rulesets/RuleSetList.vue` with entry metrics, search, namespace filter, enabled/publish/rule-presence filters, sort control, compact ruleset cards, selector summaries, and primary `管理 Rules` action.
- Kept settings and publish available as secondary actions; publish cancellation is handled without surfacing a false error.
- Hardened desktop and narrow layouts so summary cards stay compact and long selector values do not create horizontal page overflow.
- Verified with `npm run test`, `npm run type-check`, `npm run build`, and Playwright checks on `/rulesets`.
