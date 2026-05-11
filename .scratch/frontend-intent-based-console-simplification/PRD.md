# PRD: Frontend Intent-Based Console Simplification

Status: done
Created: 2026-05-10
Completed: 2026-05-10

## Problem Statement

MockServer's frontend has improved visually and the interaction hierarchy is cleaner than before, but the console still exposes more buttons, metrics, and task entries than users need at first glance. The current UI sometimes asks the user to choose between several similar entry points: global counters and page summaries repeat the same state, list rows support both row-click and explicit details actions, Workspace exposes six task tabs even though users mainly operate in three intents, and Runtime Diagnostics still shows traffic metrics before the user reaches the request problem queue.

From the user's perspective, the interface is now usable but still has residual operational noise. The next pass should not add new product capabilities. It should remove duplicate information, merge related task entries, and push secondary detail behind lightweight disclosure so the product feels simpler, calmer, and more direct.

## Solution

Reshape the UI around three core user intents:

- **Manage**: find a ruleset, inspect rules, and edit rules.
- **Debug**: inspect runtime requests, simulate a captured request, and create a rule from a problem.
- **Release**: validate, publish, inspect snapshots, and rollback when needed.

The implementation should keep the existing console/workbench visual language, route structure, backend API contracts, and data semantics. The work should be frontend-only and should focus on deleting, consolidating, or demoting UI elements:

- Global header should stop repeating page-level counts.
- Rulesets should show only the information needed to choose and manage a ruleset.
- Workspace should group detailed modes under three intent tabs.
- Rule cards should default to a compact scan view and reveal condition details only on demand.
- Runtime Diagnostics should become a debug queue first, with traffic analytics hidden until requested.

## User Stories

1. As a MockServer user, I want the global header to avoid repeating page metrics, so that it stays quiet while I work.
2. As a MockServer user, I want page-level summaries to appear only where they help the current page, so that counts do not repeat across the shell and content.
3. As a MockServer user, I want the Rulesets page to show only one summary area, so that I do not compare duplicate shown/changed/empty counts.
4. As a MockServer user, I want a ruleset row to show only the details needed for selection, so that scanning rulesets is fast.
5. As a MockServer user, I want protocol, version, and detailed rule inventory available in the details drawer, so that the list stays clean.
6. As a MockServer user, I want row click to be the details action, so that I do not need a duplicate Details button.
7. As a MockServer user, I want Workspace tasks grouped as Rules, Test, and Release, so that the workbench reflects how I think about the job.
8. As a MockServer user, I want editing to be launched from a rule rather than shown as a permanent top-level tab, so that the tab bar stays compact.
9. As a MockServer user, I want simulation and raw result to live under Test, so that testing a rule feels like one intent.
10. As a MockServer user, I want readiness, publish, snapshots, and rollback to live under Release, so that release work is grouped in one place.
11. As a MockServer user, I want rule cards to default to rule identity, state, diagnostics, and action summary, so that the rule list is readable.
12. As a MockServer user, I want priority shown as a small label instead of a large block, so that it does not dominate every card.
13. As a MockServer user, I want condition details hidden until I expand them, so that complex predicates do not crowd the list.
14. As a MockServer user, I want Runtime Diagnostics to start with recent requests, so that I can debug from concrete runtime events.
15. As a MockServer user, I want traffic analytics hidden behind Insights, so that metrics do not distract from the debug queue.
16. As a MockServer user, I want request rows to rely on row click for details, so that explicit Details buttons do not repeat the same action.
17. As a MockServer user, I want request rows to show only method, path, status, outcome, fallback, and timing by default, so that they remain compact.
18. As a MockServer user, I want debug actions to remain available in the request drawer, so that advanced actions are still one step away.
19. As a MockServer user, I want the visual language to stay consistent with previous improvements, so that the product does not feel redesigned every pass.
20. As a MockServer developer, I want the mode grouping logic to live in a small helper module, so that route modes and intent tabs remain testable.
21. As a MockServer developer, I want existing unit tests to continue covering helper behavior, so that simplification does not break routing and workflow state.
22. As a MockServer developer, I want the final implementation verified in a desktop browser, so that the PC console experience is the primary quality gate.

## Implementation Decisions

- Keep Vue 3, TypeScript, Vite, Element Plus, Pinia, Vue Router, and SCSS.
- Do not change backend APIs, persistence, rule matching, runtime metrics, publish, simulation, or rollback semantics.
- Keep the existing route query values for detailed workbench modes so deep links remain functional.
- Add intent-level workbench grouping over existing detailed modes instead of deleting the detailed modes.
- Map `inspect` and `editor` to the Rules intent, `simulate` and `result` to the Test intent, and `readiness` and `snapshots` to the Release intent.
- Keep the details drawer and existing context menus as the place for secondary actions.
- Remove duplicate global header counters rather than moving their data to a new component.
- Remove duplicate page meta where a page already has an equivalent compact summary.
- Prefer disclosure controls for non-core details instead of showing all condition/action fields in rule rows.
- Keep mobile-first redesign out of scope; only maintain basic narrow-width resilience.

## Testing Decisions

- Good tests should verify mode-to-intent mapping, default intent behavior, and existing workflow helpers, not CSS details.
- Extend existing frontend utility tests for the new workbench intent helpers.
- Existing unit tests should continue to pass.
- Type checking and production build must pass.
- Browser QA should cover desktop Rulesets, Runtime Diagnostics, and Ruleset Workspace.
- Browser QA should verify that row click still opens details and that intent switching preserves access to existing panels.

## Out of Scope

- Backend changes.
- Data model changes.
- New API endpoints.
- New persistence or history model.
- Authentication or authorization changes.
- Replacing Element Plus.
- Theme switching.
- Mobile-first redesign.
- New charts.
- Build chunk splitting.

## Further Notes

This PRD builds on the completed frontend UI/UX redesign, workflow efficiency pass, runtime debug-to-fix workflow, and interaction hierarchy redesign. The goal is stricter intentionality: every visible button and information block should earn its place on the first screen.

## Completion Notes

- Removed duplicate global header counters while keeping store refresh behavior for sidebar/page data.
- Simplified Rulesets list rows and moved protocol/version/rule inventory detail behind the existing drawer.
- Grouped Workspace primary navigation into Rules, Test, and Release intents while preserving detailed route modes.
- Compact rule cards now show priority as a small label and hide condition details behind per-card disclosure.
- Runtime Diagnostics now starts from the recent request queue, with traffic metrics available through Insights.
- Verified with unit tests, type check, production build, whitespace check, and desktop browser smoke/screenshot QA.
