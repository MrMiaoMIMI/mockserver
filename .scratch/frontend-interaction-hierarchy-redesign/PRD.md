# PRD: Frontend Interaction Hierarchy Redesign

Status: done
Created: 2026-05-10
Completed: 2026-05-10

## Problem Statement

MockServer's frontend has a cleaner visual foundation after prior UI passes, but the main pages still expose too much information at once. Rulesets, Runtime Diagnostics, and Ruleset Workspace all contain useful capabilities, but the interaction hierarchy is not strict enough: filters, metric cards, status chips, side panels, action buttons, task tabs, raw diagnostics, publish readiness, simulation, and rollback all compete for attention.

From the user's perspective, the product now looks better but can still feel visually busy. Important actions are not always visually dominant, secondary information sometimes appears before the primary task, and advanced operations are visible before the user needs them. The next improvement should make the application feel simpler, clearer, and more deliberate without removing existing capabilities.

This pass targets the PC web experience. MockServer is primarily used on desktop browsers, so mobile-specific redesign is out of scope except for keeping the layout from breaking at narrower widths.

## Solution

Redesign the frontend around a task-first workbench model:

- Each page should make one primary task obvious.
- Secondary information should move behind drawers, compact panels, disclosure controls, or contextual menus.
- Primary actions should be visually distinct and limited in number.
- High-risk actions such as publish, rollback, and delete should appear only in the relevant task context.
- Runtime diagnostics should read like a problem queue before exposing analytics.
- Ruleset authoring should keep the selected rule and active task clear at all times.
- The visual system should keep the existing clean desktop console direction while tightening status colors, spacing, card density, and action affordances.

The implementation should remain within the current Vue 3, TypeScript, Vite, Element Plus, Pinia, and SCSS stack. No backend API changes are required.

## User Stories

1. As a MockServer user, I want the Rulesets page to show the ruleset list earlier, so that I can start working without scanning many summary blocks first.
2. As a MockServer user, I want basic search and creation to stay visible, so that common entry actions are immediately available.
3. As a MockServer user, I want advanced filters hidden until I ask for them, so that the entry page stays clean.
4. As a MockServer user, I want ruleset metrics summarized compactly, so that I understand state without losing vertical space.
5. As a MockServer user, I want each ruleset row to prioritize identity, namespace, selectors, state, and the main manage action, so that list scanning is fast.
6. As a MockServer user, I want secondary ruleset actions in a drawer or menu, so that row-level actions do not clutter the list.
7. As a MockServer user, I want the Ruleset Workspace header to show only the current ruleset context and essential actions, so that the work area feels focused.
8. As a MockServer user, I want publish to live inside the readiness task, so that I do not publish before checking validation state.
9. As a MockServer user, I want workbench tasks to look like a compact desktop toolbar, so that switching between inspect, edit, simulate, readiness, snapshots, and result feels deliberate.
10. As a MockServer user, I want rule cards to expose fewer always-visible icon buttons, so that the rule list is easier to read.
11. As a MockServer user, I want destructive and less common rule actions behind a More menu, so that primary operations remain clear.
12. As a MockServer user, I want the selected rule overview to show selector, condition, and action summaries before raw details, so that I can understand behavior quickly.
13. As a MockServer user, I want Runtime Diagnostics to prioritize recent problematic requests, so that I can debug from the latest runtime signal.
14. As a MockServer user, I want runtime analytics to be available on demand, so that side panels do not dominate when they are empty or secondary.
15. As a MockServer user, I want request rows to show method, path, outcome, status, fallback reason, and primary detail action clearly, so that I do not parse scattered facts.
16. As a MockServer user, I want replay, open rules, and create rule actions to appear in the request detail drawer, so that row actions stay compact.
17. As a MockServer user, I want colors to have consistent meanings, so that blue means active or primary, green means success or published, amber means risk, and red means error.
18. As a MockServer user, I want empty states to consume less space and suggest the next useful action only when needed, so that pages do not feel padded with non-core content.
19. As a MockServer user, I want page headers and status chips to use the same compact visual language, so that the product feels unified.
20. As a MockServer user, I want the PC desktop layout to make good use of width, so that repeated work is efficient.
21. As a MockServer developer, I want shared presentation primitives to absorb interaction hierarchy decisions, so that page code stays focused on domain behavior.
22. As a MockServer developer, I want pure helper modules to continue carrying view-model logic where useful, so that behavior can be tested without browser coupling.
23. As a MockServer developer, I want the implementation to preserve existing route and API contracts, so that UI polish does not introduce backend risk.
24. As a MockServer developer, I want tests and browser QA to cover the changed desktop flows, so that the visual simplification does not regress core workflows.

## Implementation Decisions

- Keep the current frontend stack and route structure.
- Keep the existing dark sidebar and light console/workbench visual direction.
- Do not design a dedicated mobile experience in this pass. Keep basic narrow-width behavior from breaking, but optimize for PC browser usage.
- Add or refine shared presentation primitives for compact summaries, disclosure controls, task switching, row actions, and lightweight insights panels when they reduce duplication.
- Keep shared primitives presentation-focused and avoid moving page-specific business logic into them.
- Redesign the Rulesets page to move advanced filters behind a disclosure control and compress summary metrics into a compact state strip.
- Keep Rulesets row details available through the existing drawer pattern.
- Refine Ruleset Workspace so the context bar is smaller and publish is available through the readiness task rather than as a constant primary header action.
- Refine the workbench task switch into a compact desktop task bar with clear active state and no large tab cards.
- Simplify Rule Manager row actions by keeping common operations visible and moving less common or destructive operations into a row menu.
- Refine Runtime Diagnostics so recent requests dominate the layout and analytics side panels are collapsed or secondary by default.
- Keep runtime request detail drawer as the place for diagnosis, replay, open-rules, and create-rule actions.
- Tighten SCSS design tokens and shared styles for status colors, panel borders, hover states, focus states, and button density.
- Do not change rule matching, namespace fallback, simulation, publish, rollback, runtime metrics, or persistence semantics.

## Testing Decisions

- Good tests should cover helper behavior and user-visible workflow state, not CSS implementation details.
- Existing unit tests for runtime diagnostics, debug-to-fix, rule authoring, publish readiness, snapshot rollback, and entry lists should continue to pass.
- Add or adjust helper tests only where new view-model behavior is introduced.
- Type checking and production build must pass.
- Browser QA should cover desktop Rulesets, Runtime Diagnostics, and Ruleset Workspace at a PC viewport.
- Browser QA should include a narrower smoke viewport only to ensure the layout does not break, not to validate a mobile-first redesign.
- Visual QA should inspect first-screen density, action hierarchy, advanced filter disclosure, task switching, row actions, runtime request detail, and console noise.

## Out of Scope

- Backend API changes.
- Database or persistence changes.
- Rule matching behavior changes.
- Namespace fallback behavior changes.
- Simulation semantics changes.
- Publish or rollback semantics changes.
- Authentication or authorization changes.
- Replacing Element Plus.
- Adding a new charting library.
- Mobile-first redesign.
- Theme switching.
- Large navigation restructuring.

## Further Notes

This PRD builds on the prior frontend redesign, workflow-efficiency pass, runtime debug-to-fix pass, and publish/rollback workbench additions. The product already has the right capabilities; this pass should make those capabilities feel calmer, more focused, and more usable on desktop.

## Completion Notes

- Reworked Rulesets into a list-first entry page with default-hidden advanced filters, compact state summary, and row-level primary action focus.
- Simplified Ruleset Workspace by replacing the constant Publish action with a Readiness task entry and converting the task rail into a compact desktop toolbar.
- Reduced Rule Manager row noise by keeping Edit visible and moving move, priority, enable/disable, duplicate, and delete actions into a More menu.
- Refined Runtime Diagnostics into a request-first problem queue with Insights hidden by default and secondary replay/open actions kept in the detail drawer.
- Tightened shared metric card density and preserved consistent status colors across updated surfaces.
- Verified with `npm run type-check`, `npm test`, `npm run build`, `git diff --check`, Playwright screenshots, narrow-width smoke QA, and console error/warning checks.
