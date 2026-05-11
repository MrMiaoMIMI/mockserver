# Global and Rulesets noise reduction

Status: done
Type: AFK
Completed: 2026-05-10

## Parent

.scratch/frontend-intent-based-console-simplification/PRD.md

## What to build

Reduce duplicated global and Rulesets information. The global header should stop repeating draft/published/runtime counters, and the Rulesets page should use one compact summary area while keeping rows focused on choosing and managing a ruleset.

## Acceptance criteria

- [x] Global header no longer shows duplicate draft/published/runtime metric boxes.
- [x] Rulesets page title area no longer repeats shown/changed/empty when the toolbar summary already provides that information.
- [x] Ruleset rows prioritize name/id, namespace, selectors, state, and Manage rules.
- [x] Protocol, version, rule inventory, settings, and publish remain available through the details drawer or More menu.
- [x] Existing create, refresh, search, filter, details, settings, publish, and manage-rules flows still work.

## Blocked by

None - can start immediately.

## Completion notes

Implemented in `AppHeader.vue` and `RuleSetList.vue`. The shell keeps refresh behavior for shared stores, but the visible header is now action-focused. Ruleset rows are scan-first, with secondary inventory and configuration detail retained in the drawer/actions.
