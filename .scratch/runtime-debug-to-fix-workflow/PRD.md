# PRD: Runtime Debug-to-Fix Workflow

Status: done
Created: 2026-05-10
Completed: 2026-05-10

## Problem Statement

MockServer users can now see runtime traffic and replay a captured request, but the product still makes the user manually bridge the gap between "this request missed" and "which rule should I edit or create". Runtime Diagnostics exposes request facts, Ruleset Workspace exposes simulation and rule authoring, and ResultInspector exposes simulation details, but these surfaces are still loosely connected.

The highest-friction debugging loop is:

1. A runtime request falls back, misses, or returns an unexpected status.
2. The user opens the request and reads low-level fields.
3. The user must manually identify a target ruleset.
4. The user must manually copy the event into simulation.
5. The user must manually create or edit a rule from that request context.
6. The user must manually rerun simulation before publishing.

This is slow and error-prone for a developer tool whose core job is to move quickly from runtime observation to a precise mock rule fix.

## Solution

Create a Runtime Debug-to-Fix workflow that connects Runtime Diagnostics to Ruleset Workspace. A runtime request should become a portable debug payload that can open a target ruleset, prefill simulation, or seed a new rule draft. The Ruleset Workspace should explain where the payload came from, make the request fields visible, let the user run the request through draft simulation, and support creating a rule from the captured event.

The implementation should avoid backend contract changes in this pass. It should use the existing runtime metrics payload, existing simulate endpoints, existing rule editor, and existing explain rendering. Shared helper modules should own the request-to-debug-payload and request-to-rule-seed logic so it can be tested independently.

## User Stories

1. As a MockServer user, I want a runtime request to explain its likely diagnosis, so that I understand whether the issue is selector miss, rule miss, fallback, matched behavior, or runtime error.
2. As a MockServer user, I want the request detail drawer to show the next available fix actions, so that I know how to proceed from an observed failure.
3. As a MockServer user, I want to send a runtime request to Ruleset Workspace simulation, so that I can replay it against the draft ruleset without copying JSON.
4. As a MockServer user, I want to create a rule from a runtime request, so that common request fields are prefilled as match conditions.
5. As a MockServer user, I want the target ruleset selection to be automatic when possible, so that I do not need to search manually.
6. As a MockServer user, I want unmatched requests to still suggest a same-namespace HTTP ruleset when one exists, so that I can start fixing even when no ruleset matched.
7. As a MockServer user, I want the workspace to show that a debug payload was loaded from Runtime Diagnostics, so that I do not lose context after navigation.
8. As a MockServer user, I want the simulation event JSON to be prefilled from the runtime request, so that replay is one click away.
9. As a MockServer user, I want the debug payload to preserve trace ID, namespace, method, host, path, and query, so that replay matches the observed request.
10. As a MockServer user, I want replay results to remain connected to the source request ID, so that the before/after diagnosis stays clear.
11. As a MockServer user, I want generated rule drafts to avoid overfitting to trace IDs by default, so that rules apply to real traffic rather than a single debug request.
12. As a MockServer user, I want generated rule drafts to include method, host, path, and query predicates where available, so that the first draft is useful.
13. As a MockServer user, I want a generated rule draft to use a safe static response template, so that I can edit and simulate before saving.
14. As a MockServer user, I want rule authoring to show when a rule was seeded from a runtime request, so that I know why fields were prefilled.
15. As a MockServer user, I want inline validation to continue working for generated rules, so that seeded data does not bypass normal authoring safety.
16. As a MockServer user, I want explain results to read like a diagnosis path, so that I can understand which selector or condition passed or missed.
17. As a MockServer user, I want browser navigation from runtime request to ruleset/rule context to preserve selected rule when the runtime request already matched a rule.
18. As a MockServer user, I want the workflow to keep existing create, edit, simulate, validate, publish, snapshot, and rollback behavior intact, so that the new workflow does not remove existing control.
19. As a MockServer developer, I want request-to-rule-seed logic in a tested helper module, so that UI components do not duplicate domain logic.
20. As a MockServer developer, I want storage of debug payloads to be narrow and temporary, so that runtime request data does not become hidden application state.
21. As a MockServer developer, I want the implementation to remain frontend-only unless a backend gap is proven, so that this iteration stays small and shippable.
22. As a MockServer developer, I want tests to cover helper behavior, so that future UI changes do not break debug payload generation.
23. As a MockServer developer, I want browser QA to cover runtime-to-workspace navigation, so that the workflow is verified as a user journey.
24. As a MockServer developer, I want console checks to stay clean, so that new routing and session storage logic does not create noisy warnings.

## Implementation Decisions

- Keep the current Vue 3, TypeScript, Vite, Element Plus, Pinia, Vue Router, Axios, and SCSS stack.
- Do not add backend API endpoints for this pass. Use runtime metrics records and existing simulate endpoints.
- Add a tested frontend utility module for Runtime Debug-to-Fix payloads.
- Store debug payloads in `sessionStorage` with a short payload ID passed through route query parameters.
- Keep session payloads narrow: source request facts, target ruleset ID, requested action, and captured event.
- Derive target ruleset from the runtime record when `ruleset_id` is present; otherwise fall back to the first enabled draft in the same namespace and protocol.
- Add Runtime Diagnostics actions for using a request as simulation input and creating a rule from the request when a target ruleset can be inferred.
- In Ruleset Workspace, consume the debug payload after route load and prefill simulation or editor creation according to the requested action.
- Extend Rule Editor creation mode to accept an optional seed rule while retaining normal validation and save behavior.
- Generated rules should include request method, host, path, and first query values as structured predicates where available.
- Generated rules should not include trace ID as a default condition, but the trace ID should remain visible in the source debug context.
- Generated rules should use a static response action template with HTTP 200 and a small JSON body that references the runtime source.
- Enhance the workspace simulation panel with a source banner when an event was loaded from Runtime Diagnostics.
- Keep existing route contracts and add only query parameters for workflow context.
- Keep ResultInspector compatible with current simulation payloads while improving diagnosis-oriented presentation where practical.

## Testing Decisions

- Good tests should verify external helper behavior: target ruleset inference, payload shape, event labeling, and generated rule seed content.
- Do not test CSS implementation details.
- Add unit tests for the new debug-to-fix helper module under the existing frontend utility test pattern.
- Existing runtime diagnostics, diagnostics, rule authoring, and rule form adapter tests should continue to pass.
- Frontend type checking and production build should pass.
- Browser QA should cover:
  - Runtime request detail drawer.
  - Use-as-simulation navigation into Ruleset Workspace.
  - Create-rule-from-request navigation into Rule Editor.
  - Simulation execution from loaded runtime event.
  - Narrow viewport smoke check.

## Out of Scope

- Backend explain payload schema changes.
- New persistent storage schema.
- Authentication or user-level audit for debug payloads.
- Multi-step guided publish wizard.
- Automatic rule saving without user confirmation.
- Automatic publish after rule creation.
- Global diff of runtime request against every ruleset.
- Long-term debug history.
- Replacing the current Rule Editor.
- Replacing the current ResultInspector raw JSON support.

## Further Notes

This PRD builds on the completed UI redesign and workflow-efficiency pass. The next meaningful product step is to reduce the distance between runtime observation and a concrete rule fix.

## Completion Notes

- Added a tested Runtime Debug-to-Fix helper module for target ruleset inference, debug payload creation, diagnosis copy, session payload routing, and runtime-derived rule seeding.
- Extended Runtime Diagnostics request detail with diagnosis summary, inferred target, use-as-simulation, and create-rule actions.
- Extended Ruleset Workspace to consume debug payloads, show runtime source context, prefill simulation, and seed create-rule authoring.
- Enhanced simulation output with a compact diagnosis path before detailed selector and condition traces.
- Verified with `npm run type-check`, `npm test`, `npm run build`, `git diff --check`, and Playwright browser QA screenshots under `output/playwright/debug-to-fix-*.png`.
