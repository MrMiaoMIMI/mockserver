# PRD: Rule Workspace Interaction Simplification

Status: done
Created: 2026-05-11
Completed: 2026-05-11

## Problem Statement

The Ruleset Workspace page is visually cleaner than earlier versions, but the rule authoring flow still exposes too many secondary controls and technical details on the first screen. Users see duplicated navigation and refresh actions, a tall filter block that pushes rules down, rule card menu items that duplicate edit behavior, a create-rule form that asks for both Rule ID and Name, an unclear "active task" label, and JSON areas that are hard to read when inspecting or editing request and result payloads.

From the user's perspective, the workspace still feels more like an internal data editor than a focused mock-rule console. The next pass should remove low-value buttons, compress filters, keep rule card actions purposeful, make rule creation name-first, rename the right-side work area with domain language, and make JSON inspection/editing comfortable on a PC browser.

## Solution

Refine the Ruleset Workspace around two primary surfaces:

- **Rule Manager**: a compact scan-and-select list for finding rules, opening edit, ordering, duplicating, and deleting rules.
- **Rule Workspace**: the right-side focused work area for inspecting, testing, releasing, and editing the selected rule.

The solution should keep existing backend contracts, simulation behavior, publish behavior, route structure, and rule semantics. The implementation is frontend-focused:

- Remove redundant context actions that duplicate sidebar/global navigation.
- Collapse non-essential rule filters behind a single Filters popover while keeping active filter state visible.
- Remove quick actions that are already represented by the editor.
- Make new-rule creation name-first and generate Rule ID automatically unless the user enters advanced raw JSON.
- Rename "active task" to "Rule Workspace" and clarify the current panel title.
- Upgrade shared JSON editing/viewing to a high-contrast, readable tool surface with copy, format, line numbers, and larger usable space.

## User Stories

1. As a MockServer user, I want the Ruleset Workspace top bar to avoid duplicate navigation actions, so that only meaningful workspace actions remain visible.
2. As a MockServer user, I want list navigation to happen through the sidebar or browser back, so that the ruleset context menu is not cluttered.
3. As a MockServer user, I want data refresh to happen through the global refresh button, so that I do not see multiple refresh entry points.
4. As a MockServer user, I want ruleset details to remain available through a simple details toggle, so that secondary metadata is one step away.
5. As a MockServer user, I want the Rule Manager filter area to consume less vertical space, so that more rules fit on screen.
6. As a MockServer user, I want search to remain immediately visible, so that I can quickly locate a rule by id, name, condition, action, or status.
7. As a MockServer user, I want enabled/action/diagnostic filters grouped under one Filters control, so that advanced filtering does not dominate the rule list.
8. As a MockServer user, I want active filters shown as small chips, so that I can see why a rule list is narrowed.
9. As a MockServer user, I want a clear way to reset all filters, so that I can recover from an empty filtered result quickly.
10. As a MockServer user, I want status counts summarized compactly, so that I still understand the stack health without losing space.
11. As a MockServer user, I want rule cards to expose only actions that are commonly needed outside editing, so that the menu is easier to choose from.
12. As a MockServer user, I want priority changes to happen in Edit, so that ordering and identity edits are handled in one place.
13. As a MockServer user, I want enabled/disabled changes to happen in Edit, so that state edits are intentional and reviewed with other rule properties.
14. As a MockServer user, I want the card menu to focus on move, duplicate, and delete, so that destructive and structural actions are easy to recognize.
15. As a MockServer user, I want to create a rule by giving it a name, so that I do not need to invent a technical id.
16. As a MockServer user, I want the Rule ID to be generated from the name, so that the system keeps stable identifiers without making me manage them manually.
17. As a MockServer user, I want generated Rule IDs to avoid duplicates, so that saving a new rule does not fail unexpectedly.
18. As a MockServer user, I want the generated Rule ID to be visible as secondary text, so that I can inspect it when needed.
19. As a MockServer user, I want existing rules to keep their current Rule ID, so that simulation, diff, and publish behavior remain stable.
20. As a MockServer user, I want raw JSON mode to remain available for advanced edits, so that power users can still control the full payload.
21. As a MockServer user, I want the right-side panel to be named "Rule Workspace", so that I understand it is the focused work area for the selected rule.
22. As a MockServer user, I want the current panel title to describe the actual workflow, so that I know whether I am inspecting, testing, editing, or releasing.
23. As a MockServer user, I want JSON request payloads to be readable at a glance, so that I can safely edit simulation input.
24. As a MockServer user, I want JSON result payloads to be readable at a glance, so that I can inspect simulation and release responses.
25. As a MockServer user, I want JSON blocks to support copy and format actions, so that I can move data between tools quickly.
26. As a MockServer user, I want JSON blocks to use high-contrast code styling, so that braces, indentation, and long strings are not visually washed out.
27. As a MockServer user, I want JSON blocks to have line numbers, so that I can discuss and locate payload sections more precisely.
28. As a MockServer user, I want JSON blocks to have enough height for PC usage, so that request and result payloads do not feel cramped.
29. As a MockServer developer, I want generated rule id logic isolated in a helper, so that it can be unit tested without mounting the editor.
30. As a MockServer developer, I want JSON formatting behavior centralized in the shared JSON component, so that all JSON surfaces improve together.
31. As a MockServer developer, I want existing route and backend contracts unchanged, so that this UX pass is low-risk.
32. As a MockServer developer, I want final verification to include browser QA on the target page, so that the visual and interaction changes are validated in the actual console.

## Implementation Decisions

- Keep this pass frontend-only unless an existing frontend helper needs testable extraction.
- Keep the current Vue 3, TypeScript, Vite, Element Plus, Pinia, Vue Router, and SCSS stack.
- Do not change backend APIs, rule storage, publish semantics, simulation semantics, snapshot behavior, or route query modes.
- Replace the ruleset context More menu with a direct details toggle because Back and Refresh duplicate existing shell controls.
- Keep Settings visible because it opens ruleset-level configuration and is not duplicated in the workspace body.
- Keep Rule Manager search inline and move enabled/action/diagnostic filters into a popover.
- Show active filters as compact chips only when non-default filters are active.
- Keep card-level Edit as the primary rule action.
- Keep card-level More for move up, move down, duplicate, and delete.
- Remove card-level Set Priority and Enable/Disable because they are editor-owned fields.
- Generate new rule ids from the rule name when creating through the form.
- Preserve existing rule ids during edit and in raw JSON advanced mode.
- Keep Rule ID validation and duplicate prevention in the form adapter because the backend still requires stable ids.
- Rename the right-side label from "active task" to "Rule Workspace".
- Keep the intent tabs Rules, Test, and Release from the prior optimization pass.
- Upgrade the shared JSON editor with toolbar actions, line number gutter, clearer code colors, copy, format, and an expanded presentation mode.
- Reuse the shared JSON editor for Test event JSON, Raw result JSON, Rule raw JSON, condition JSON, and authoring preview JSON.

## Testing Decisions

- Unit tests should cover generated rule id behavior, duplicate avoidance, and existing locked-id behavior.
- Existing workbench intent tests should continue to pass.
- UI tests should focus on visible behavior: redundant actions gone, filter popover available, card menu simplified, name-first create form, Rule Workspace label visible, and JSON editor controls visible.
- Do not test CSS internals directly.
- Type checking, unit tests, production build, and whitespace checks must pass.
- Browser QA should cover `http://localhost:6173/rulesets/http-default-test1-b03f4c4f/rules` on desktop because this is the primary usage context.

## Out of Scope

- Backend schema or API changes.
- Authentication and permission changes.
- Mobile-first redesign.
- Replacing Element Plus.
- Drag-and-drop rule ordering.
- Full Monaco/CodeMirror integration.
- New charting or analytics.
- Build chunk splitting.

## Further Notes

This PRD continues the current console/workbench visual direction. The goal is not to add more capability; it is to make existing capability easier to operate by removing redundant controls, demoting technical details, and improving payload readability.

## Completion Notes

- Removed redundant ruleset context Back/Refresh actions from the normal workspace surface and replaced the More menu with a direct Details toggle.
- Renamed the right-side work area to Rule Workspace while keeping the current workflow title and selected-rule context visible.
- Compressed Rule Manager filters into a search-first toolbar plus Filters popover, active chips, reset, and compact counts.
- Reduced rule card More actions to move, duplicate, and delete; priority and enabled state remain editable in the rule editor.
- Changed create-rule identity UX to name-first with generated Rule ID preview and tested duplicate-safe ID generation.
- Upgraded the shared JSON editor/viewer with line numbers, higher contrast, copy, format, and expanded mode.
- Verified with frontend type check, frontend unit tests, production build, backend Go tests, whitespace check, and browser QA on the target page.
