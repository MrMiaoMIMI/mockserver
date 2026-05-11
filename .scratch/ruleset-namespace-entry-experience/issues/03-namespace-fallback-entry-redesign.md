# Namespace Fallback Entry Redesign

Status: done
Type: AFK

## Parent

.scratch/ruleset-namespace-entry-experience/PRD.md

## What to build

Redesign Namespace List as a fallback policy entry surface. Users should see ruleset miss and rule miss behavior, forward versus response policy, and ruleset usage before editing a namespace.

## Acceptance criteria

- [x] Namespace List shows high-signal metrics for total, shown, forward policies, response policies, and used namespaces.
- [x] Users can search by namespace ID, name, description, fallback type, related ruleset ID, and related ruleset name.
- [x] Users can filter by ruleset miss type, rule miss type, and usage state.
- [x] Users can sort by namespace ID, usage count, ruleset miss policy, and rule miss policy.
- [x] Each namespace row/card shows name, ID, description, usage count, related rulesets, ruleset miss policy, and rule miss policy.
- [x] Forward fallback and response fallback are visually distinct and readable.
- [x] Existing create/edit fallback behavior and default forward behavior are preserved.
- [x] Frontend type-check, tests, and build pass.

## Blocked by

- .scratch/ruleset-namespace-entry-experience/issues/01-entry-view-models.md

## Comments

- Redesigned `web/src/views/namespaces/NamespaceList.vue` with fallback metrics, search, usage/fallback filters, sorting, namespace usage counts, related ruleset links, and side-by-side miss policy summaries.
- Reused and simplified `FallbackSummary` through the shared fallback summary helper so list and dialog previews use the same policy labels.
- Preserved create/edit fallback payload behavior and default forward form state.
- Added a policy preview to the edit dialog and fixed dialog layering by appending entry dialogs to body with responsive widths.
- Verified with `npm run test`, `npm run type-check`, `npm run build`, and Playwright checks on `/namespaces`.
