# PRD: Ruleset Workspace Refactor

Status: done
Created: 2026-05-07

## Problem Statement

The current ruleset rules page at `/rulesets/:id/rules` does not give MockServer users a clear, efficient workspace for authoring and diagnosing a ruleset. The page technically exposes the required operations: inspect draft metadata, add and edit rules, update priority, enable or disable rules, validate, publish, simulate, inspect snapshots, preview rollback, and execute rollback. However, the experience feels assembled from independent widgets rather than designed around the user's actual workflow.

The main page currently mixes four high-complexity jobs into one split layout:

- ruleset-level orientation
- rule stack authoring
- simulation and result diagnosis
- published snapshot and rollback operations

This creates several user-visible problems:

- The rule stack only gives shallow condition and action summaries, so users must open a modal to understand or edit most rule behavior.
- The rule editor modal removes context from the rule stack and forces identity, condition, action, and raw JSON concerns into one large form.
- The simulate dock competes with rule editing for attention, but the default event payload is static and may not match the current ruleset's namespace, selector, or first rule.
- Validation, simulation, publish, and rollback results all collapse into a generic result tab and raw JSON output, so users must interpret operational outcomes manually.
- Snapshot and rollback flows are hidden behind a tab and do not provide enough current-versus-target context before rollback.
- Page-level state orchestration is concentrated in the view component, while rule form serialization and summary generation are embedded inside the rule manager component.
- Backend APIs already expose most primitives, but the frontend has to compose draft, published, snapshots, validation, and simulation state manually; if this remains noisy after frontend refactor, a backend workspace summary endpoint may be justified.

The user wants this module reviewed, redesigned, and turned into concrete optimization requirements. The goal is a better ruleset workspace, not a small visual polish pass.

## Solution

Refactor the ruleset rules page into a focused Ruleset Workspace that makes the user's workflow explicit: orient, select a rule, edit or inspect it, simulate traffic, review diagnostics, then publish or rollback with confidence.

The redesigned workspace should prioritize three persistent areas:

- A compact ruleset header that shows draft identity, namespace, selector, version, published state, validation state, and primary actions.
- A rule stack that supports scanning, selecting, reordering, enabling, disabling, duplicating, deleting, and understanding rules without opening every rule.
- A workbench area that changes mode based on the active task: rule detail/editing, simulation, validation result, publish result, snapshot history, and rollback preview.

The design should move away from a generic "right dock with tabs" toward a task-oriented workspace. Rule editing should preserve context, simulation should derive better defaults from the current ruleset, and diagnostics should be structured for decision-making instead of only dumping raw JSON.

Implementation should be modular. The refactor should extract deep modules that own state orchestration and rule form mapping behind simple interfaces, so the page component no longer has to coordinate every backend call and every presentation concern directly.

Backend changes are allowed when they reduce frontend complexity or expose domain state the frontend cannot reliably derive. The current backend is already close to sufficient, so backend work should be scoped and justified. A likely backend addition is a workspace summary API that returns draft, current published snapshot summary, snapshot count/list summary, validation summary, and optional rule metrics in one response. No storage schema change is expected.

## User Stories

1. As a MockServer user, I want the ruleset workspace to show the ruleset's identity, namespace, selector, version, draft state, and published state at a glance, so that I know exactly what I am editing.
2. As a MockServer user, I want the workspace title and header actions to stay oriented around the current ruleset, so that I do not lose context while editing rules.
3. As a MockServer user, I want the rule stack to be the primary authoring surface, so that rules are easy to scan, compare, and operate.
4. As a MockServer user, I want each rule row to show priority, enabled state, rule ID, name, condition summary, action summary, and last relevant diagnostic state, so that I can understand rules before opening an editor.
5. As a MockServer user, I want to select a rule and inspect its details inline, so that I can keep the rule stack visible while reading the selected rule.
6. As a MockServer user, I want to edit a selected rule in a contextual panel or drawer, so that editing does not feel disconnected from the rule stack.
7. As a MockServer user, I want identity, condition, and action editing to be separated into clear sections or tabs, so that a large rule form is not overwhelming.
8. As a MockServer user, I want condition editing to support predicate, ALL, ANY, NOT, CEL, and raw JSON modes clearly, so that I can choose the right level of control.
9. As a MockServer user, I want action editing to show only fields relevant to the selected action type, so that static, template, CEL, sequence, and webhook actions are easier to configure.
10. As a MockServer user, I want sequence responses to be edited as ordered steps, so that I do not need to hand-edit a large JSON array for common sequence changes.
11. As a MockServer user, I want webhook response configuration to clearly show method, URL, timeout, and headers, so that webhook behavior is easy to audit.
12. As a MockServer user, I want JSON editing to remain available for advanced cases, so that the UI does not block complex rule definitions.
13. As a MockServer user, I want rule form validation errors to point to the exact field or section, so that I can fix invalid input quickly.
14. As a MockServer user, I want backend validation issues to map back to the relevant rule or ruleset section, so that compile problems are actionable.
15. As a MockServer user, I want priority changes to be easy and safe, so that I can reorder rule matching behavior without guessing.
16. As a MockServer user, I want to duplicate an existing rule, so that I can create similar rules without rebuilding condition and action details manually.
17. As a MockServer user, I want destructive actions such as delete and rollback to clearly state their target, so that I do not accidentally remove or restore the wrong rule behavior.
18. As a MockServer user, I want simulation defaults to be derived from the current ruleset namespace, protocol, selector hosts, selector path prefixes, and first useful rule condition, so that the initial simulation is meaningful.
19. As a MockServer user, I want to switch simulation target between draft and published while keeping the event payload stable, so that I can compare what will change before publishing.
20. As a MockServer user, I want simulation results to show matched ruleset, matched rule, fallback state, fallback reason, response status, and candidate rules, so that I can diagnose routing quickly.
21. As a MockServer user, I want condition explanations to be shown as a readable trace, so that I can understand why a rule matched or missed.
22. As a MockServer user, I want selector checks to be visible when a ruleset is not selected, so that I can distinguish selector miss from rule condition miss.
23. As a MockServer user, I want raw JSON results to remain available after structured summaries, so that I can debug edge cases without losing detail.
24. As a MockServer user, I want validation and simulation results to persist in the workspace until superseded, so that I can compare edits against the latest diagnostic output.
25. As a MockServer user, I want publish to show validation status and publish result clearly, so that I know whether the draft became the published version.
26. As a MockServer user, I want the workspace to show whether the draft differs from the current published version, so that I know if publishing is necessary.
27. As a MockServer user, I want snapshot history to show snapshot ID, version, published time, operator, reason, and source snapshot when available, so that rollback choices are understandable.
28. As a MockServer user, I want rollback preview to show current versus target differences before rollback, so that I can judge the impact safely.
29. As a MockServer user, I want rollback preview simulation to reuse the same event payload as normal simulation, so that rollback impact can be evaluated against real traffic examples.
30. As a MockServer user, I want narrower viewports to stack the workspace into a usable order, so that rule editing and diagnosis remain possible without horizontal scrolling.
31. As a MockServer user, I want the workspace to keep the existing console/workbench visual direction, so that the refactor feels like a product improvement rather than a new unrelated UI.
32. As a MockServer user, I want loading, saving, validating, simulating, publishing, and rollback states to be visible per operation, so that I know which action is in progress.
33. As a MockServer user, I want errors from API calls to appear next to the relevant workflow, so that a failed simulation or failed publish does not look like a generic page problem.
34. As a MockServer developer, I want page-level orchestration extracted into a workspace state module, so that the view component stays small and predictable.
35. As a MockServer developer, I want rule form serialization and deserialization extracted into a testable adapter, so that UI form shape can evolve without corrupting API rule payloads.
36. As a MockServer developer, I want condition summary and action summary generation centralized, so that list rows, detail panels, and diagnostics use consistent language.
37. As a MockServer developer, I want simulation event defaults generated by a dedicated helper, so that defaults can be tested against selectors and common condition shapes.
38. As a MockServer developer, I want result parsing and diagnostic view models extracted from the display component, so that validation, simulation, and rollback results can be tested as domain transformations.
39. As a MockServer developer, I want backend additions limited to stable domain contracts, so that the API does not become tightly coupled to a single visual layout.
40. As a MockServer maintainer, I want the refactor split into tracer-bullet implementation slices, so that the workspace can improve incrementally without a long-running rewrite.

## Implementation Decisions

- Keep the existing route contract for ruleset workspace navigation.
- Keep Vue 3, TypeScript, Vite, Element Plus, Pinia, and SCSS as the frontend stack.
- Treat the page as a Ruleset Workspace, not just a "rules list" page.
- Redesign the information architecture around persistent ruleset context, selectable rule stack, and task-oriented workbench modes.
- Replace the current generic right-side tab dock with clearer workbench modes for rule details, simulation, diagnostics, snapshots, and rollback.
- Keep rule operations immediate-save for this PRD unless implementation discovers that draft staging is required. Do not introduce a full unsaved local draft system in the first refactor.
- Add duplicate-rule as a frontend workflow using existing add-rule semantics unless backend support is needed for ID generation or validation.
- Continue using existing admin endpoints for rule add, update, delete, enable, disable, priority, validate, publish, simulate, snapshots, rollback preview, and rollback where they are sufficient.
- Consider adding a backend workspace summary endpoint only if it materially simplifies the frontend. The endpoint should return stable domain data such as draft, current published summary, snapshot summaries, validation summary, and optional runtime metrics by rule.
- Do not add backend endpoints that exist only to mirror a visual component hierarchy.
- Preserve current rule, condition, action, ruleset, namespace, published snapshot, simulation, validation, and rollback JSON contracts unless a narrowly scoped contract change is justified.
- Generate default simulation events from current ruleset context instead of using a static event. The generator should prefer the ruleset namespace and protocol, then selector host/path hints, then simple request condition hints when safe.
- Keep raw JSON event editing available, but pair it with structured controls for common fields such as method, namespace, host, path, query, and headers.
- Create a testable workspace state module that owns loading draft context, published state, snapshots, active rule, active workbench mode, operation loading states, and current diagnostic result.
- Create a testable rule form adapter that maps API `Rule` objects into editable form state and maps form state back to valid API `Rule` payloads.
- Create condition and action summary helpers that produce concise, stable summaries for rule cards and detail headers.
- Create a simulation event default helper that turns a ruleset and optional selected rule into a useful default event payload.
- Create a diagnostic view model helper that turns validation, simulation, publish, rollback preview, and rollback responses into structured panels before raw JSON.
- Keep `SelectorValueGroup` as the reusable pattern for compact selector values with inspect-on-demand behavior.
- Reuse the existing JSON editor for advanced payload, body, and raw result editing, but place it inside clearer task contexts.
- Preserve existing namespace fallback behavior and runtime matching semantics.
- Preserve existing publish and rollback audit behavior.
- Preserve existing admin authentication and operator header behavior.
- Preserve existing database schema unless backend summary work proves that a derived metric requires a new persistent field. This is not expected.
- Responsive behavior should stack the workspace in this order: ruleset context, primary actions, rule stack, active workbench.
- Avoid page sections styled as nested decorative cards. Use panels only where they frame real tools or repeated items.
- Use stable dimensions for rule rows, icon buttons, tabs, counters, and compact controls so hover states and content changes do not shift the layout.

## Testing Decisions

- Good tests should verify external behavior and stable domain transformations, not the exact internal component structure.
- Frontend build and type checks must pass after the refactor.
- If frontend unit test infrastructure is introduced or already available at implementation time, test the rule form adapter, summary helpers, simulation event default helper, diagnostic view model helper, and workspace state module.
- Rule form adapter tests should cover every action type: static response, template response, CEL response, sequence response, and webhook response.
- Rule form adapter tests should cover condition modes: predicate, ALL, ANY, NOT, CEL expression, and raw JSON.
- Simulation event default tests should cover deriving namespace, protocol, host, path, method, and simple condition values from a ruleset.
- Diagnostic view model tests should cover validation pass/fail, simulation matched/missed, fallback result, selector miss explanation, rule miss explanation, rollback diff changed/same, and raw JSON fallback.
- Backend tests should be added only for backend contract changes.
- If a workspace summary API is added, cover it through controller-level tests that assert the response includes draft, current published summary, snapshot summaries, validation summary, and stable absent-state behavior.
- If backend simulation or validation contracts are extended, add controller tests and service tests that assert old semantics remain unchanged while new fields are present.
- Preserve existing controller tests that cover rule management, validation, publish, simulate, snapshot, rollback preview, and rollback.
- Use Chrome or browser-based manual QA for the redesigned workspace at desktop width and narrower viewport width.
- Manual QA should cover: empty draft, one-rule draft, multi-rule draft, long rule IDs, long selector values, disabled rule, invalid rule, publish success, simulation miss, simulation match, snapshot list, rollback preview, and rollback confirmation.
- Manual QA should verify that the default simulation event for the target ruleset no longer points at an unrelated namespace.
- Manual QA should verify that raw JSON remains accessible in diagnostics and advanced editors.
- Manual QA should verify that long code-like values truncate, wrap, or open inspection affordances predictably.
- `git diff --check` should pass.

## Out of Scope

- Changing runtime request matching semantics.
- Changing rule priority semantics.
- Changing namespace fallback semantics.
- Changing published snapshot storage semantics.
- Replacing Element Plus.
- Introducing a new frontend framework or state management library.
- Building a full visual no-code rule engine beyond the current condition and action capabilities.
- Adding multi-user collaboration.
- Adding permissions, roles, or audit UI beyond showing existing publish and rollback audit fields.
- Adding a persistent unsaved local draft model unless later implementation proves it is necessary.
- Adding production-grade drag-and-drop reordering in the first pass if simple priority controls are sufficient.
- Adding new runtime protocols beyond the current HTTP-focused workspace.
- Adding database schema migrations unless a backend summary decision explicitly requires one.
- Implementing the refactor in this PRD step. This PRD defines the work; implementation should be split into issues afterward.

## Further Notes

Review observations from the current implementation:

- The page component currently handles route state, draft loading, published-state lookup, snapshot loading, validation, publish, simulation, rollback preview, rollback execution, result JSON formatting, and default event creation.
- The rule manager currently combines rule list rendering, modal visibility, form state, condition mode state, action form state, rule serialization, rule deserialization, JSON parsing, summary generation, and backend mutation calls.
- The current static default event uses `namespace: "default"`, while the reviewed target ruleset is in `my-namespace-15a80b12`. This can make first-run simulation misleading.
- Result inspection currently gives a few structured panels but still depends heavily on raw JSON for useful diagnosis.
- Backend APIs already expose most needed operations. The first implementation pass should try to simplify the frontend with client-side deep modules before adding backend contracts.
- If backend support is added, prefer a stable domain-level workspace summary over multiple UI-specific endpoints.
- This project is greenfield, so the refactor does not need to preserve the old layout or old component boundaries when a better design is available.
