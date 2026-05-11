# PRD: Frontend Workflow Efficiency

Status: done
Created: 2026-05-10
Completed: 2026-05-10

## Problem Statement

MockServer's frontend now has a cleaner visual foundation, but the product workflows still carry too much page-local structure. Lists, cards, filters, state chips, metric summaries, task switches, and diagnostic rows repeat similar UI ideas with independent markup and CSS. This makes future UI changes harder and leaves the user experience more page-oriented than task-oriented.

The highest friction remains in three daily workflows:

- Finding the right ruleset among many entries and understanding its core state.
- Moving through the rule management loop from inspect to edit, simulate, validate, publish, snapshot, and raw result.
- Diagnosing runtime behavior and jumping from a request or replay result back to the owning ruleset/rule.

The user wants the next optimization pass to improve practical usability, not just theme color.

## Solution

Introduce a small set of shared UI primitives and use them to reshape the main workflows. Rulesets should become easier to scan at scale with a list-first surface and a detail drawer. Ruleset Workspace should become more task-flow oriented while preserving all existing capabilities. Runtime Diagnostics should expose a clearer request diagnosis loop with inspectable request detail and direct navigation to the owning rule context.

The implementation should keep the current backend API contracts and frontend stack. Improvements should be made in the Vue frontend through reusable components, page layout changes, state handling, and browser-verified UX.

## User Stories

1. As a MockServer user, I want repeated controls to look and behave consistently, so that I do not relearn each page.
2. As a MockServer user, I want metrics to use the same visual language across pages, so that summary data is easy to compare.
3. As a MockServer user, I want state chips to communicate the same tones everywhere, so that enabled, draft, changed, warning, and error states are predictable.
4. As a MockServer user, I want segmented filters to have consistent keyboard and visual behavior, so that filtering feels reliable.
5. As a MockServer user, I want key/value details to be readable and compact, so that IDs, namespace, protocol, version, and counts can be scanned quickly.
6. As a MockServer user, I want Rulesets to scale beyond a few cards, so that a larger mock configuration remains manageable.
7. As a MockServer user, I want the Rulesets page to prioritize identity, state, namespace, selectors, rule count, and main action, so that I can choose the right ruleset quickly.
8. As a MockServer user, I want secondary ruleset details available in a drawer, so that the list remains clean while detailed data remains inspectable.
9. As a MockServer user, I want ruleset settings and publish actions to remain close to the ruleset, so that I can act without losing context.
10. As a MockServer user, I want the rule workspace to show the next task clearly, so that I can move from rule selection to edit, simulate, validate, and publish.
11. As a MockServer user, I want the workbench switch to feel like a task rail, so that advanced panels do not feel like unrelated tabs.
12. As a MockServer user, I want the selected rule context to be obvious while editing or simulating, so that I do not modify the wrong rule.
13. As a MockServer user, I want runtime requests to be inspectable, so that I can understand a request without reading raw JSON first.
14. As a MockServer user, I want replay output to stay connected to the source request, so that diagnosis has a clear before/after path.
15. As a MockServer user, I want a runtime request to navigate directly to the owning ruleset and selected rule, so that fixing behavior is fast.
16. As a MockServer user, I want empty states to point to the next useful action, so that I know what to do when no data is present.
17. As a MockServer user, I want narrow viewport behavior to remain usable, so that demos and quick checks do not require a large screen.
18. As a MockServer developer, I want shared UI primitives to reduce page-local style duplication, so that future frontend changes are lower risk.
19. As a MockServer developer, I want component interfaces to be small and stable, so that they are deep enough to reuse without carrying page-specific logic.
20. As a MockServer developer, I want existing tests and browser QA to cover the changed surfaces, so that workflow improvements do not regress existing behavior.

## Implementation Decisions

- Keep Vue 3, TypeScript, Vite, Element Plus, Pinia, Vue Router, Axios, and SCSS.
- Do not change backend APIs, route contracts, rule matching behavior, namespace fallback behavior, publish behavior, simulation behavior, rollback behavior, or persistence.
- Add shared frontend primitives for metrics, state chips, segmented filters, key/value grids, task rails, and empty/action panels where they remove meaningful duplication.
- Keep the shared components presentation-focused and avoid moving page business logic into generic UI components.
- Redesign the Rulesets page into a list-first surface with detail drawer support.
- Keep selector values compact and inspectable through the existing selector value affordance.
- Keep the existing settings dialog for ruleset configuration, but expose it from the new list/detail workflow.
- Refine Ruleset Workspace around a task rail and selected-rule context without rewriting rule editing, simulation, readiness, snapshots, or result internals.
- Add runtime request detail inspection in the frontend and keep replay connected to the selected request.
- Preserve current browser QA practice with desktop and narrow viewport checks.

## Testing Decisions

- Good tests should verify user-visible behavior and pure helper behavior, not CSS implementation details.
- Existing unit tests should continue to pass.
- Frontend build/type checking should pass.
- Browser QA should cover Rulesets, Ruleset Workspace, Runtime Diagnostics, and a narrow viewport.
- Visual QA should specifically inspect list density, drawer content, segmented controls, task rail behavior, selected rule context, request detail inspection, and replay result context.
- Planning artifacts should be marked complete only after build, tests, and browser QA pass.

## Out of Scope

- Backend API changes.
- New storage schema.
- Rule matching changes.
- Namespace fallback behavior changes.
- Publishing or rollback semantic changes.
- Replacing Element Plus.
- Adding a theme switcher.
- Adding a new charting library.
- Rewriting the whole frontend state model.
- Replacing the route structure.

## Further Notes

This PRD builds on the completed frontend UI/UX redesign. The visual theme is now good enough; this pass should improve task efficiency, reduce duplication, and make the core MockServer loops feel more like a focused developer tool.

## Completion Notes

- Added shared frontend primitives for metric cards, state chips, segmented filters, key/value grids, and task rails.
- Reworked Rulesets into a list-first workflow with a detail drawer and retained create, settings, publish, and manage-rules actions.
- Refined Ruleset Workspace with a task rail and shared Rule Manager filters/status chips.
- Improved Runtime Diagnostics with inspectable request detail, source-linked replay, and ruleset/rule navigation context.
- Verified with `npm run type-check`, `npm run build`, `npm test`, `git diff --check`, and Playwright browser QA screenshots under `output/playwright/`.
