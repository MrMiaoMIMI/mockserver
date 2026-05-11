# Entry View Models For Rulesets And Namespaces

Status: done
Type: AFK

## Parent

.scratch/ruleset-namespace-entry-experience/PRD.md

## What to build

Create reusable entry-page view models for Ruleset List and Namespace List. The view models should derive publish state, fallback state, usage counts, search text, filters, sort order, and display summaries from existing admin API data.

## Acceptance criteria

- [x] Ruleset entry view model derives published, draft-only, changed, enabled, disabled, empty-rule, selector count, and searchable text states.
- [x] Namespace entry view model derives fallback labels, forward/response tone, usage count, related ruleset IDs, and searchable text states.
- [x] Filters and sorting are pure functions that can be tested without Vue components.
- [x] Unit tests cover ruleset publish state, ruleset filters, ruleset sorting, namespace fallback summaries, namespace usage, namespace filters, and namespace sorting.
- [x] Existing backend API contracts remain unchanged.
- [x] Frontend type-check and tests pass.

## Blocked by

None - can start immediately.

## Comments

- Added `web/src/utils/entryLists.ts` for ruleset and namespace entry rows, metrics, fallback summaries, filters, and sorting.
- Added `web/src/utils/__tests__/entryLists.test.ts` covering publish-state derivation, ruleset filters/sorting/metrics, fallback summaries, namespace usage, namespace filters/sorting/metrics.
- Verified with `npm run test` and `npm run type-check`.
