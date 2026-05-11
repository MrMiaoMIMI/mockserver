# Rulesets entry information hierarchy

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-interaction-hierarchy-redesign/PRD.md

## What to build

Reshape the Rulesets entry page so the primary list appears earlier and the page feels less crowded. Keep search, refresh, and create immediately available, but move advanced filters behind a compact disclosure affordance. Replace the large metric strip with a compact summary that preserves the important state counts without consuming excessive vertical space.

## Acceptance criteria

- [x] Rulesets page first screen prioritizes search, primary actions, and the ruleset list.
- [x] Advanced filters are hidden by default and can be opened without losing state.
- [x] Summary metrics are compact and use consistent status language.
- [x] Ruleset rows keep identity, namespace, selectors, state, and Manage rules as the primary row information.
- [x] Secondary ruleset actions remain available through drawer or menu interactions.
- [x] Existing Rulesets create, settings, publish, and manage-rules flows still work.

## Completion notes

- Collapsed advanced namespace/status filters behind a Filters button.
- Replaced large metric cards with compact state summary pills.
- Reduced row actions to Manage rules plus More while preserving details, settings, and publish.

## Blocked by

None - can start immediately.
