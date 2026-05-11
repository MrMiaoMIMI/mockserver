# PRD: Default Namespace Forward Fallback

Status: needs-triage
Created: 2026-05-07

## Problem Statement

MockServer already supports namespace-level fallback policies for two runtime miss cases:

- `ruleset_miss`: no published ruleset selector matches the runtime request.
- `rule_miss`: a published ruleset is selected, but no rule inside that ruleset matches the runtime request.

Today, the built-in default namespace configuration returns synthetic 404 JSON responses for both miss cases. The frontend's new namespace form also defaults both fallback actions to response mode. This makes MockServer behave like a closed mock endpoint by default: any uncovered request is stopped at MockServer instead of continuing to the real upstream target.

The desired default behavior is different. For the `default` namespace, both ruleset miss and rule miss should forward the original request. The same default should apply when creating other new namespaces, so users can add targeted mock rules without having to manually reconfigure pass-through behavior each time.

## Solution

Change the namespace fallback defaults so both `ruleset_miss_action` and `rule_miss_action` use forward mode by default. Runtime fallback should continue to forward the original method, scheme, host, path, query, headers, and body using the existing forward fallback behavior. The default timeout should remain consistent with the current frontend forward fallback default unless implementation chooses a shared constant.

After the change:

- The `default` namespace starts with both miss actions set to forward.
- If a runtime request in `default` misses all rulesets, MockServer forwards the original request.
- If a runtime request in `default` selects a ruleset but misses all rules, MockServer forwards the original request.
- Newly created namespaces default both miss actions to forward.
- Users can still explicitly change either fallback action to a synthetic response when they want MockServer to stop unmatched traffic.
- The existing fallback response capability remains available for custom namespace policies.

## User Stories

1. As a MockServer user, I want uncovered traffic in the `default` namespace to pass through to the original upstream, so that adding mock rules does not break endpoints I have not mocked yet.
2. As a MockServer user, I want a ruleset miss in the `default` namespace to forward the original request, so that requests outside configured selectors still behave like the real service.
3. As a MockServer user, I want a rule miss in the `default` namespace to forward the original request, so that partially mocked rulesets do not block unmatched paths or payloads.
4. As a MockServer user, I want newly created namespaces to default to forward fallback, so that every namespace starts as pass-through until I deliberately override it.
5. As a MockServer user, I want to preserve the option to return a configured response on miss, so that I can still build strict mock-only namespaces when needed.
6. As a MockServer user, I want the forwarded request to preserve the original method, so that `GET`, `POST`, `PUT`, `DELETE`, and other methods reach upstream correctly.
7. As a MockServer user, I want the forwarded request to preserve the original request path, so that upstream receives the same application route the client intended.
8. As a MockServer user, I want the forwarded request to preserve the original query string, so that upstream behavior depending on query parameters remains unchanged.
9. As a MockServer user, I want the forwarded request to preserve the original request body, so that write and search APIs still receive their payloads.
10. As a MockServer user, I want the forwarded request to preserve useful headers, so that upstream auth, content negotiation, and trace behavior can keep working.
11. As a MockServer user, I want hop-by-hop headers to remain excluded from forwarding, so that proxy-specific headers do not corrupt the upstream call.
12. As a MockServer user, I want MockServer to expose whether a fallback happened, so that I can tell the difference between a mock hit and pass-through fallback.
13. As a MockServer user, I want ruleset miss and rule miss to remain distinguishable, so that I can diagnose whether the selector or the rule condition caused pass-through.
14. As a MockServer operator, I want the default namespace shown in the admin UI to display forward fallback for both miss cases, so that the UI reflects the runtime behavior.
15. As a MockServer operator, I want the new namespace dialog to open with forward selected for both miss cases, so that the safest default is visible before saving.
16. As a MockServer operator, I want the namespace list filters and summaries to continue working with forward and response policies, so that I can audit fallback configuration across namespaces.
17. As a MockServer developer, I want backend default namespace construction to be the source of truth for default fallback actions, so that runtime fallback and admin APIs stay consistent.
18. As a MockServer developer, I want frontend defaults to match backend defaults, so that the UI does not save response fallback unintentionally.
19. As a MockServer developer, I want tests for both miss reasons under default forward behavior, so that future changes do not accidentally restore synthetic 404 defaults.
20. As a MockServer developer, I want existing custom response fallback tests to remain valid, so that explicit response fallback behavior is not regressed.
21. As a MockServer developer, I want existing custom forward fallback tests to remain valid, so that original-request forwarding remains covered by behavior tests.
22. As a MockServer developer, I want no new storage schema just to change defaults, so that the feature stays scoped to namespace policy defaults and runtime behavior.
23. As a MockServer developer, I want no compatibility adapter for old defaults, so that this greenfield project can directly converge on the desired behavior.
24. As a MockServer tester, I want a simple end-to-end case where no rulesets exist in `default`, so that ruleset miss forwarding is proven.
25. As a MockServer tester, I want a simple end-to-end case where a published ruleset exists but no rule matches, so that rule miss forwarding is proven.

## Implementation Decisions

- The default namespace fallback policy should be forward for both ruleset miss and rule miss.
- Newly created namespaces should default to forward for both ruleset miss and rule miss.
- The existing response fallback type remains supported and user-selectable.
- The existing forward fallback execution path should be reused rather than introducing a second proxy mechanism.
- The default forward fallback should forward the original request target derived from the runtime request, including scheme, original host, path, query, method, body, and non-hop-by-hop headers.
- The fallback reason should remain `ruleset_miss` when no published ruleset selector is selected.
- The fallback reason should remain `rule_miss` when a published ruleset is selected but no rule matches.
- Runtime responses produced by fallback should continue to carry the fallback indicator header.
- Backend namespace default construction should be updated so any missing or auto-created default namespace uses forward fallback actions.
- Backend namespace normalization should fill missing fallback actions with the new forward defaults when appropriate for new namespace creation, instead of forcing clients to provide response fallback defaults.
- Frontend namespace creation defaults should be updated so both fallback editors start in forward mode.
- Frontend ruleset settings namespace options should not synthesize old response fallback defaults for typed namespace values.
- Existing explicit namespace policies should continue to round-trip through admin APIs without being rewritten unintentionally.
- No database schema change is required. The namespace JSON shape already supports both response and forward fallback actions.
- No runtime URL contract change is required. Runtime still uses the existing namespace-scoped HTTP entrypoint.
- No new admin endpoint is required. Existing namespace create, update, get, and list endpoints should expose the new defaults through their current contracts.
- No compatibility layer is required for old default values because this project is still greenfield.

## Testing Decisions

- Good tests for this feature should assert external behavior through runtime HTTP calls and admin namespace APIs, not internal helper implementation details.
- Add or update backend tests so a runtime request in the default namespace with no published matching ruleset forwards to an upstream test server.
- Add or update backend tests so a runtime request in the default namespace with a selected ruleset but no matching rule forwards to an upstream test server.
- Assert that forwarded fallback preserves path and query string.
- Assert that forwarded fallback returns the upstream status, headers, and JSON body through the runtime response.
- Assert that the runtime response marks fallback and keeps the correct fallback reason for ruleset miss and rule miss.
- Keep tests that prove explicitly configured response fallback still returns the configured response for both ruleset miss and rule miss.
- Keep tests that prove explicitly configured forward fallback for a custom namespace uses the original request target.
- Add or update namespace service tests so the default namespace seed uses forward fallback for both miss policies.
- Add or update admin API tests so creating a namespace without explicit fallback actions, if supported by the implementation, yields forward defaults.
- Add or update frontend unit tests if the repo already has frontend test infrastructure for namespace form defaults. If not, validate through focused build/type checks and manual UI inspection.
- Run focused Go tests for controller/runtime/service behavior after backend changes.
- Run the frontend build after UI default changes, because the namespace form and type definitions are part of the user-facing workflow.

## Out of Scope

- Changing how ruleset selector matching works.
- Changing how rule condition matching works.
- Changing rule priority or ruleset strictness ordering.
- Adding new fallback action types beyond response and forward.
- Adding upstream target configuration independent of the original request.
- Adding retry, circuit breaker, or advanced proxy behavior for forwarded fallbacks.
- Adding authentication transformation for forwarded requests.
- Changing runtime URL structure.
- Changing database schema.
- Migrating historical namespace rows in persistent databases unless implementation work finds that current local development data must be reset or updated manually.
- Reworking the namespace UI beyond the default fallback selection and any copy needed to reflect the new default.

## Further Notes

Current code already has most of the runtime machinery needed for this behavior: namespace fallback actions support response and forward types, runtime miss handling distinguishes ruleset miss from rule miss, and forward fallback already builds an upstream request from the original runtime request. The main product change is to make forward the default policy consistently across backend defaults, default namespace bootstrap behavior, and frontend namespace creation.
