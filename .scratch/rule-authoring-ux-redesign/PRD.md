# PRD: Rule Authoring UX Redesign

Status: done
Created: 2026-05-08

## Problem Statement

MockServer now has stronger ruleset management, namespace fallback configuration, publish readiness, and runtime diagnostics. The next bottleneck is rule authoring itself. Users can observe what happened at runtime and jump back to the owning rule, but creating or fixing a rule still requires too much knowledge of the underlying rule payload shape.

The current editor exposes tree, CEL, raw JSON, action types, headers, bodies, sequence steps, and webhook fields, but it does not guide users through common mock behavior. A user who wants to match a method and path, add a query/header/body condition, return a static response, or quickly simulate the draft must still understand fields such as request.path, request.headers, JSON value encoding, and when each action field matters.

This makes the rule management loop slower than it should be: diagnose a runtime request, open the rule, edit conditions or response behavior, validate, simulate, then save. The editor should make that loop obvious, readable, and low-error.

## Solution

Redesign the rule authoring experience around a guided builder while preserving advanced access.

The condition editor should promote common request predicates through reusable presets such as method, path, query, header, body field, host, and trace ID. Users should be able to add these predicates without memorizing field names. Predicate rows should display readable labels, concise operator wording, JSON-safe value editing, and inline validation.

The rule editor should present a compact authoring overview, live validation, condition builder, action builder, advanced JSON, and draft simulation as one connected workflow. Static response authoring should be especially clear because it is the common case. Template, CEL, sequence, and webhook actions should stay available but should explain what the selected action will do.

Users should be able to generate a simulation event from the current rule and ruleset selector, run a draft simulation before saving where possible, and inspect the result in the existing structured result inspector. Raw JSON editing remains available for advanced users, but it should not be the default path.

## User Stories

1. As a MockServer user, I want to create a rule from common request templates, so that I do not need to memorize internal field names.
2. As a MockServer user, I want method matching to be a first-class control, so that I can quickly match GET, POST, PUT, DELETE, and other methods.
3. As a MockServer user, I want path matching presets, so that exact path, prefix, suffix, contains, and regex use cases are easy to configure.
4. As a MockServer user, I want query parameter matching presets, so that I can match request query values without writing request.query paths manually.
5. As a MockServer user, I want header matching presets, so that I can match request headers without guessing the field syntax.
6. As a MockServer user, I want body field matching presets, so that I can target common JSON payload fields safely.
7. As a MockServer user, I want each condition row to show readable labels and operator names, so that I can scan what a rule does.
8. As a MockServer user, I want condition values to be JSON-safe but friendly, so that strings, booleans, numbers, arrays, and objects do not require trial and error.
9. As a MockServer user, I want inline validation for incomplete predicates, so that I can fix mistakes before saving or publishing.
10. As a MockServer user, I want grouped ALL, ANY, and NOT conditions to remain available, so that complex rules can still be expressed.
11. As a MockServer user, I want CEL expressions to remain available, so that advanced matching is still possible.
12. As a MockServer user, I want raw condition JSON to remain available, so that advanced payloads can still be edited directly.
13. As a MockServer user, I want static response authoring to be clear, so that I can set status, headers, and body without understanding every action type.
14. As a MockServer user, I want action types to explain their behavior, so that I can choose static, template, CEL, sequence, or webhook responses correctly.
15. As a MockServer user, I want response headers and body inputs to validate JSON live, so that bad response payloads are caught early.
16. As a MockServer user, I want long paths, headers, body values, and JSON snippets to be readable and copyable, so that debugging large rules is practical.
17. As a MockServer user, I want a concise authoring overview, so that I can understand the current rule before saving it.
18. As a MockServer user, I want visible readiness for identity, condition, action, and raw JSON sections, so that I know which part needs attention.
19. As a MockServer user, I want to simulate the current draft rule from the editor, so that I can verify behavior before leaving the authoring flow.
20. As a MockServer user, I want the simulation event to be generated from the ruleset selector and current condition, so that the preview starts with realistic traffic.
21. As a MockServer user, I want to edit the generated simulation event when needed, so that I can test edge cases.
22. As a MockServer user, I want simulation results to use the existing result inspector, so that diagnostics are consistent with the rules workspace.
23. As a MockServer user, I want the editor to work on narrow viewports, so that controls do not overlap or become unreadable.
24. As a MockServer developer, I want rule authoring helpers to be testable outside the Vue component, so that validation and generated simulation behavior remain stable.
25. As a MockServer maintainer, I want the existing rule payload shape to stay compatible, so that backend matching behavior does not change.

## Implementation Decisions

- Treat this as an authoring workflow redesign, not a theme redesign.
- Keep Vue 3, TypeScript, Vite, Element Plus, Pinia, SCSS, and the existing console/workbench visual direction.
- Preserve the existing backend rule payload shape for rules, conditions, actions, simulation requests, and validation responses.
- Add a testable frontend authoring helper that knows condition presets, readable operator labels, section readiness, JSON validation, and simulation event suggestions.
- Keep the condition tree editor as the core builder, but add preset-driven predicate creation and friendlier predicate editing.
- Keep ALL, ANY, NOT, CEL, and raw JSON pathways available for advanced rules.
- Make static response the default action path and keep other action types available through a clearly described chooser.
- Use the existing ruleset workspace store methods for draft simulation and save operations.
- Use the existing structured result inspector for authoring preview results.
- Do not add a new route; this work stays inside the ruleset workspace authoring panel.
- Do not change matching semantics, action semantics, namespace fallback semantics, or publish semantics.
- Keep advanced raw JSON editing synchronized with the guided form.

## Testing Decisions

- Good tests should verify externally observable authoring behavior: preset condition output, readable summaries, inline validation, action readiness, generated simulation events, and rule payload compatibility.
- Unit tests should cover the new authoring helper because it is the deep module behind the UI.
- Existing rule form adapter tests should be extended where payload conversion or validation behavior changes.
- Frontend tests should verify that rule form conversion still produces the same API payload shape for existing rule types.
- Browser checks should cover desktop and narrow viewport authoring flows, condition preset controls, action chooser readability, simulation preview, and no obvious overlap or clipping.
- Verification should include frontend tests, type-check, production build, Go tests, `git diff --check`, and browser QA.

## Out of Scope

- Changing backend condition or action semantics.
- Persisting editor drafts outside the existing ruleset draft model.
- Replacing CEL with a custom expression language.
- Building a full rule-template marketplace.
- Reworking runtime diagnostics again.
- Reworking namespace fallback pages again.
- Adding authentication or permission changes.
- Persisting simulation history.

## Further Notes

This PRD follows the runtime diagnostics work. The desired loop is now: observe runtime behavior, jump to the owning rule, author a fix with guided controls, simulate the draft, then save or publish with confidence.

## Completion Notes

- Implemented a guided rule authoring workflow focused on condition presets, action clarity, section readiness, and draft preview simulation.
- Added `draft_override` support for draft simulation so users can test the current editor form before saving it.
- Verified with frontend unit tests, type-check, production build, backend Go tests, `git diff --check`, and Playwright browser QA on desktop plus narrow viewport.
