# PRD: Frontend Readability And Typography Polish

Status: needs-triage
Created: 2026-05-07

## Problem Statement

MockServer's frontend has a polished console-style layout, but some information-dense fields are currently too small to read comfortably. The issue is most visible in operational surfaces such as ruleset cards, namespace cards, selector values, rule summaries, simulation results, rollback diffs, dashboard rows, and compact metadata chips.

The current design system defines a small base type scale and uses the smallest text token broadly across code values, labels, chips, summaries, and secondary fields. This makes important operational information hard to scan, especially when the user needs to compare rulesets, inspect rule conditions, read request paths, review fallback policies, or diagnose why a simulation missed.

The user wants a noticeably better frontend reading experience. The goal is not to redesign the product again, but to tune typography, spacing, and dense field presentation so the UI remains compact while becoming comfortable for daily use.

## Solution

Improve the frontend's readability by revising the typography scale and applying it consistently across information-dense components. Important values should be readable without leaning in, code-like fields should have enough size and line height, labels should remain secondary without becoming illegible, and compact chips should stop using the smallest text size for content that users must actually inspect.

The solution should preserve the existing MockServer console/workbench direction, Element Plus integration, Vue 3 architecture, and page structure. The implementation should focus on practical readability improvements:

- Raise the effective body and field-value sizes to a more comfortable baseline.
- Reduce overuse of the smallest text token in operational data.
- Make code values, selector values, fallback summaries, rule condition summaries, and result rows easier to read.
- Keep labels and metadata visually secondary, but not too small.
- Improve vertical rhythm where slightly larger text needs more breathing room.
- Keep dense pages scannable without turning the UI into a spacious marketing layout.
- Verify the result in desktop and smaller viewport layouts so text does not overlap, clip awkwardly, or force unnecessary horizontal scrolling.

## User Stories

1. As a MockServer user, I want important field values to be readable at normal viewing distance, so that I can operate the tool without eye strain.
2. As a MockServer user, I want ruleset IDs, namespace IDs, paths, hosts, and rule IDs to be easier to read, so that I can quickly identify the object I am editing.
3. As a MockServer user, I want selector values to remain compact but readable, so that I can compare request routing conditions without opening every popover.
4. As a MockServer user, I want full selector value popovers to use readable code text, so that long hosts and paths can be inspected comfortably.
5. As a MockServer user, I want rule condition summaries to be readable, so that I can understand a rule before opening the edit dialog.
6. As a MockServer user, I want rule action summaries to be readable, so that I can distinguish response, sequence, and webhook behavior at a glance.
7. As a MockServer user, I want simulation summary fields to be readable, so that I can diagnose matched and missed traffic quickly.
8. As a MockServer user, I want validation issue rows to be readable, so that I can fix invalid rulesets without copying the raw JSON elsewhere.
9. As a MockServer user, I want rollback diff rows to be readable, so that I can judge whether a rollback is safe.
10. As a MockServer user, I want JSON editor text to stay comfortable for reading and editing request events, rule bodies, and raw results.
11. As a MockServer user, I want dashboard metric labels and keys to be readable, so that runtime metrics can be scanned without squinting.
12. As a MockServer user, I want namespace fallback summaries to be readable, so that I can understand response versus forward behavior quickly.
13. As a MockServer user, I want filter controls and segmented buttons to remain compact but legible, so that filtering does not become a guessing exercise.
14. As a MockServer user, I want small metadata chips to use readable text, so that status, counts, namespace, and publish state remain useful.
15. As a MockServer user, I want the UI to keep a professional operations-console feel, so that readability improvements do not make the app feel oversized or childish.
16. As a MockServer user, I want long code-like values to wrap, truncate, or expose full text predictably, so that larger fonts do not break layouts.
17. As a MockServer user, I want the layout to remain stable when text gets larger, so that hover states, buttons, cards, and panels do not jump around.
18. As a MockServer user, I want the UI to remain usable on narrower screens, so that larger type does not cause overlapping or clipped content.
19. As a MockServer developer, I want readability changes to be driven by shared tokens, so that future pages inherit sane defaults.
20. As a MockServer developer, I want repeated dense-value rendering to be handled through reusable components or shared styles, so that fixes do not drift page by page.
21. As a MockServer developer, I want Element Plus overrides to align with the app's type scale, so that form labels, inputs, tables, dialogs, and buttons feel consistent.
22. As a MockServer developer, I want code review to identify remaining too-small text uses, so that the first pass does not miss hidden surfaces.
23. As a MockServer developer, I want build and type checks to pass after style changes, so that visual polish does not introduce frontend regressions.
24. As a MockServer reviewer, I want before/after screenshots for key pages, so that the readability improvement is visible and not just token churn.
25. As a MockServer reviewer, I want the implementation to avoid unrelated visual redesign, so that the scope stays focused on readability.

## Implementation Decisions

- Treat this as a typography and dense-field readability pass, not a full product redesign.
- Keep the current Vue 3, Vite, Element Plus, Pinia, and SCSS architecture.
- Use the existing design system as the main control point for typography tokens, spacing, and shared component defaults.
- Revisit the current type scale where the base size is small and `xs` is used for many operational values.
- Prefer increasing readable value text to a comfortable size over only increasing labels.
- Keep labels visually secondary through weight, color, casing, and hierarchy, not through making them too small to read.
- Use the monospace font for code-like values, but make code values large enough for daily inspection.
- Avoid scaling font size with viewport width.
- Preserve zero letter spacing.
- Preserve the existing color direction unless contrast problems are discovered during implementation.
- Update Element Plus overrides so form labels, inputs, textareas, table text, dialogs, buttons, and empty states inherit the improved scale.
- Update shared utility classes such as code chips and compact metadata patterns so repeated surfaces improve together.
- Update dense repeated surfaces: ruleset cards, namespace cards, ruleset workspace strip, rule stack rows, result inspector panels, selector value chips/popovers, fallback summaries, and dashboard metric/detail rows.
- Keep card and panel density reasonable by adjusting spacing and min widths where needed after text increases.
- Use reusable shared styles or reusable components for dense key/value and code-value presentation when that removes duplication.
- Do not introduce a new design system or external UI library.
- Do not change backend APIs, rule matching behavior, namespace fallback behavior, storage, or runtime semantics.
- Do not hide data to make layouts look cleaner. If data is too long, provide predictable truncation, wrapping, popover inspection, or copy behavior.

## Testing Decisions

- Good tests for this work should verify externally visible frontend behavior and layout quality, not the exact internal CSS implementation.
- The frontend build should pass after the typography changes.
- Type checking should pass through the existing build command or dedicated type-check command.
- Use browser screenshots to review the key pages at desktop size: dashboard, rulesets list, ruleset workspace, and namespaces list.
- Use browser screenshots or responsive inspection for a narrower viewport to ensure larger text does not overlap, clip, or destroy scanability.
- Inspect the ruleset list with long selector values to ensure chips remain readable and full values remain inspectable.
- Inspect the ruleset workspace with rule summaries, JSON editor, simulation results, and rollback result panels.
- Inspect namespace cards and fallback summaries to ensure fallback policies are readable.
- Inspect dialogs for ruleset settings, rule editing, and namespace editing to ensure labels, controls, and textarea content remain comfortable.
- Verify long code-like values do not resize their containers unexpectedly on hover.
- Verify buttons and icon controls still have stable dimensions after text token changes.
- Run a search over frontend styles for remaining uses of the smallest text token and confirm each is intentional.
- Manual visual QA is required because this is primarily a readability and layout-quality change.

## Out of Scope

- A full frontend redesign or new visual identity.
- Backend API changes.
- Runtime matching behavior changes.
- Namespace fallback behavior changes.
- Database schema changes.
- New feature workflows beyond improving readability of the existing pages and controls.
- Adding user-configurable font size preferences.
- Adding accessibility settings panels.
- Replacing Element Plus.
- Adding a new charting library or visual analytics system.
- Rewriting the frontend state management.
- Reworking routing or navigation structure.

## Further Notes

The current frontend already has the right high-level structure for this work: shared design tokens, Element Plus overrides, common page containers, JSON editor, selector value component, result inspector, ruleset and namespace pages, and rule management components. The likely highest-impact changes are to the shared typography tokens and the repeated dense-value components. The implementation should start there, then audit page-specific overrides that still force `xs` text onto values users need to read.
