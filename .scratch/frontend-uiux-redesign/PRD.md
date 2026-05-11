# PRD: Frontend UI/UX Redesign

Status: done
Created: 2026-05-10

## Problem Statement

MockServer's frontend currently feels visually heavy and hard to scan. The dark console theme uses very similar values for page background, panels, controls, table areas, and card surfaces, so boundaries are often unclear. Green is used as the brand color, primary action, selected state, hover state, and success indicator, which makes important signals blend together instead of communicating distinct meaning.

Users need to operate MockServer as a daily developer console: find rulesets, inspect selectors, manage rules, understand namespace fallback policy, and diagnose runtime behavior. The current UI can technically expose these workflows, but the visual hierarchy makes common decisions slower than necessary.

## Solution

Redesign the frontend into a clean, light, professional developer console with a restrained dark navigation rail. The interface should be simple, fresh, spacious where it matters, and compact where users need repeated operational scanning.

The redesign should:

- Use a light content canvas with clear white surfaces and strong but restrained borders.
- Keep a dark sidebar only as a stable navigation anchor.
- Use blue for primary actions and selected navigation/filter state.
- Reserve green for success/enabled/running, amber for warnings/draft changes, red for danger/errors, and cyan/violet only for secondary analytics accents.
- Make cards, tables, forms, segmented controls, dialogs, popovers, and workbench panels feel like one coherent system.
- Reduce visible non-core detail by making secondary data collapsible, summarized, or inspectable on demand.
- Preserve the existing frontend stack, route structure, stores, backend APIs, and MockServer semantics.

## User Stories

1. As a MockServer user, I want page regions to have clear boundaries, so that I can understand where controls, summaries, and content begin and end.
2. As a MockServer user, I want primary actions to stand out from status indicators, so that I can quickly identify what to click next.
3. As a MockServer user, I want enabled and success states to remain green, so that they do not get confused with generic selected filters.
4. As a MockServer user, I want warnings, draft-only states, and validation concerns to use a distinct warning color, so that operational risk is obvious.
5. As a MockServer user, I want list pages to prioritize core fields, so that I can scan rulesets and namespaces without reading every detail.
6. As a MockServer user, I want detailed selectors, fallback policies, and raw payloads to remain available on demand, so that the cleaner interface does not hide important debugging information.
7. As a MockServer user, I want the Rulesets page to make search, filtering, publishing, and rule management easy to find, so that I can move from overview to action quickly.
8. As a MockServer user, I want the ruleset workspace to separate rule browsing from editing, simulation, readiness, snapshots, and raw results, so that I can stay oriented during complex work.
9. As a MockServer user, I want Namespaces to present fallback policy clearly, so that I can understand what happens when no ruleset or rule matches.
10. As a MockServer user, I want Runtime Diagnostics to highlight traffic health before lower-priority details, so that I can diagnose live behavior faster.
11. As a MockServer user, I want empty and low-data states to be calm and readable, so that the app does not feel noisy when little data exists.
12. As a MockServer user, I want the UI to remain usable on a narrow viewport, so that I can inspect or demo it without desktop-only assumptions.
13. As a MockServer user, I want buttons, tabs, filters, cards, and dialogs to share the same visual language, so that the product feels consistent.
14. As a MockServer developer, I want the redesign to be driven by shared design tokens, so that future pages inherit the new visual language.
15. As a MockServer developer, I want Element Plus overrides to align with the app design system, so that imported components do not look bolted on.
16. As a MockServer developer, I want repeated page patterns to use shared utility classes where practical, so that page-specific drift is reduced.
17. As a MockServer developer, I want the implementation to avoid backend changes, so that the UI/UX work stays low-risk.
18. As a MockServer reviewer, I want browser screenshots across key routes, so that the visual improvement can be verified directly.
19. As a MockServer reviewer, I want build and type checks to pass, so that the redesign does not introduce frontend regressions.
20. As a future contributor, I want planning issues and completion notes to reflect the shipped work, so that later UI work can build from the new baseline.

## Implementation Decisions

- Treat this as a complete frontend visual and interaction redesign, not a compatibility-constrained polish pass.
- Keep Vue 3, TypeScript, Vite, Element Plus, Pinia, Vue Router, Axios, and SCSS.
- Keep the current backend API contracts and route structure.
- Do not change rule matching behavior, namespace fallback behavior, publish semantics, rollback semantics, simulation semantics, or storage schema.
- Move the default app canvas to a light theme.
- Keep a dark left sidebar for product identity and navigation.
- Use blue as the application primary color.
- Reserve green for enabled/success/running states.
- Use shared tokens for surfaces, text, borders, shadows, control fills, state colors, sidebar colors, header colors, and focus rings.
- Align Element Plus buttons, inputs, select menus, tables, dialogs, dropdowns, popovers, tags, alerts, tabs, and empty states with those tokens.
- Reduce noisy decorative backgrounds and grid effects in the main content area.
- Make page headers compact and action-oriented.
- Keep metadata chips useful but visually secondary.
- Keep card border radius at 8px or less.
- Keep detail-heavy content inspectable through popovers, drawers, collapsible sections, workbench modes, or existing detail panels rather than showing all raw detail by default.
- Update Rulesets, Ruleset workspace, Namespaces, Runtime Diagnostics, shared layout, shared components, and repeated dense data patterns.
- Use browser QA on desktop and a narrow viewport as part of the implementation, not as a separate future task.

## Testing Decisions

- Good tests for this work verify externally visible behavior: pages render, core workflows remain accessible, data remains visible or inspectable, and layout does not overlap on desktop or narrow viewport.
- Run the existing frontend build, which includes type checking.
- Run existing frontend unit tests if they remain applicable.
- Use browser screenshots for Rulesets, Ruleset workspace, Namespaces, Runtime Diagnostics, and a narrow Rulesets viewport.
- Inspect visual states manually for contrast, hierarchy, button prominence, segmented controls, cards, dialogs, and workbench panels.
- Verify no backend tests are required unless frontend changes reveal an API issue.
- Update PRD and issue statuses after implementation and QA.

## Out of Scope

- Backend API changes.
- Database schema changes.
- Rule matching changes.
- Namespace fallback behavior changes.
- Publish, rollback, and simulation semantic changes.
- New runtime analytics features.
- Replacing Element Plus.
- Adding a theme switcher or user-configurable color settings.
- Adding a new charting library.
- Rewriting frontend state management.
- Replacing the existing route structure.

## Further Notes

The current frontend already has useful product structure: a sidebar shell, common page container, entry pages, a ruleset workbench, namespace fallback editors, runtime diagnostics, shared design tokens, and Element Plus overrides. The redesign should reuse that structure while making the visual system clearer, calmer, and more operationally legible.

## Completion Notes

- Completed on 2026-05-10.
- Implemented a light main console theme with a dark navigation rail, blue primary/selected states, green success/enabled states, clearer surfaces, stronger borders, and reduced dark-grid visual noise.
- Applied the visual system across Rulesets, Ruleset workspace, Namespaces, Runtime Diagnostics, shared layout, shared Element Plus overrides, and repeated ruleset components.
- Verified with `npm run build`, `npm test`, `git diff --check`, and browser QA screenshots for desktop Rulesets, Ruleset workspace, Namespaces, Runtime Diagnostics, and narrow Rulesets viewport.
