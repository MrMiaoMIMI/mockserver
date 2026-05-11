# PRD: Namespace Policy List Redesign

Status: done
Created: 2026-05-11
Completed: 2026-05-11

## Problem Statement

The Namespaces page is currently presented as a card grid. That layout looks consistent with some earlier console surfaces, but it does not match the way users evaluate namespace fallback policy. A namespace is primarily a policy record: users need to scan identity, usage, ruleset-miss fallback, rule-miss fallback, and related rulesets side by side.

The current card design spreads that information vertically, repeats high-level summary data, and includes a `View rules` action that jumps to the first related ruleset. This is confusing when a namespace is linked to more than one ruleset because the button implies a namespace-level rules view, while the actual behavior opens one arbitrary ruleset's rules page.

## Solution

Redesign the Namespaces page as a compact policy list:

- Show namespaces in a table-like list optimized for scanning and comparison.
- Keep the namespace name and ID visible, with the name as the primary label.
- Show usage, ruleset-miss fallback, rule-miss fallback, and linked rulesets as first-class list columns.
- Remove the ambiguous `View rules` action.
- Let each linked ruleset be the explicit navigation target.
- Collapse secondary details behind compact controls or popovers instead of rendering large card blocks.
- Keep `Edit policy` as the primary row action.
- Preserve the existing backend API and namespace data model.

## User Stories

1. As a MockServer user, I want namespaces shown in a compact list, so that I can compare fallback policies quickly.
2. As a MockServer user, I want namespace name and ID shown together, so that I can recognize the namespace while still seeing the technical identifier.
3. As a MockServer user, I want ruleset and rule counts visible in each row, so that I can understand whether a namespace is actively used.
4. As a MockServer user, I want ruleset-miss policy visible in each row, so that I can see what happens when no ruleset selector matches.
5. As a MockServer user, I want rule-miss policy visible in each row, so that I can see what happens when a ruleset matches but no rule matches.
6. As a MockServer user, I want linked rulesets shown as explicit targets, so that I know exactly where a navigation will go.
7. As a MockServer user, I want the page to avoid a `View rules` button that chooses the first ruleset, so that I am not surprised by an arbitrary destination.
8. As a MockServer user, I want multiple linked rulesets to be discoverable without taking much space, so that dense namespaces remain readable.
9. As a MockServer user, I want unused namespaces to remain easy to spot, so that cleanup candidates are visible.
10. As a MockServer user, I want the existing filters and sorting to keep working, so that I can narrow the namespace list by usage and fallback type.
11. As a MockServer user, I want high-level counts to remain visible but not dominate the page, so that the primary list has more vertical space.
12. As a MockServer user, I want the edit action to stay near each namespace, so that policy edits are easy to start.
13. As a MockServer user, I want fallback details to stay readable, so that I do not need to open raw JSON to understand behavior.
14. As a MockServer developer, I want namespace list behavior to stay covered by view-model tests, so that filtering, sorting, and usage summaries do not regress.
15. As a MockServer developer, I want the page to reuse the existing frontend request and store contracts, so that this redesign stays UI-focused.
16. As a MockServer developer, I want browser QA on the real Namespaces page, so that list density and linked ruleset navigation are verified in the running app.

## Implementation Decisions

- Keep the existing namespace API and store contract unchanged.
- Treat namespace as a policy-list item, not a card-oriented content tile.
- Use the existing namespace entry view model for usage, fallback summary, filtering, sorting, and searchable text.
- Add a small linked-ruleset presentation model only if it reduces repeated logic in the page.
- Replace the grid of namespace cards with a list surface that has stable columns for identity, usage, fallback policy, linked rulesets, and actions.
- Keep fallback summaries compact by default; detailed fallback editing remains in the existing create/edit dialog.
- Remove the `View rules` action because namespace does not own a single rules page.
- Preserve per-ruleset navigation through linked ruleset controls.
- Keep page-level filters and create/edit actions intact.
- Reduce duplicated metric presentation by making summary counts secondary to the list.

## Testing Decisions

- Unit tests should cover namespace entry behavior through the existing view-model helpers rather than CSS details.
- Tests should verify multiple linked rulesets are retained in the namespace usage summary and remain searchable.
- Type checking should catch the page refactor and component prop changes.
- Browser QA should verify the real `/namespaces` page renders as a list, has no `View rules` button, and allows linked ruleset navigation to an explicit ruleset.
- Build verification should ensure the refactor remains production-safe.

## Out of Scope

- Backend schema or API changes.
- New namespace detail route.
- Bulk namespace operations.
- Deleting namespaces.
- Changing fallback policy semantics.
- Mobile-first redesign. MockServer is primarily used on desktop web.

## Further Notes

This PRD continues the UI simplification direction already applied to the ruleset and rule workbench: keep core information visible, remove ambiguous actions, and move secondary information behind simple, explicit affordances.

## Completion Notes

- The Namespaces page now renders namespace policies as a compact list instead of a card grid.
- The list shows namespace identity, usage, ruleset-miss fallback, rule-miss fallback, linked rulesets, and edit action as stable columns.
- The ambiguous `View rules` action was removed.
- Linked ruleset navigation is now explicit through individual ruleset controls.
- Added namespace usage test coverage for multiple linked rulesets.
- Verified with frontend type checking, targeted namespace tests, full frontend tests, production build, `git diff --check`, and browser QA on `/namespaces`.
