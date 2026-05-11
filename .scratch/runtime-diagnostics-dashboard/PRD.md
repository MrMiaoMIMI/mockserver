# PRD: Runtime Diagnostics Dashboard

Status: done
Created: 2026-05-08

## Problem Statement

The rule configuration workflow has become stronger: users can find rulesets, manage rules, understand namespace fallback, and inspect publish readiness. The missing piece is the runtime feedback loop. After users publish or test traffic, the current Runtime Board only shows aggregate totals and hit rankings. It does not help users answer the operational questions that matter during debugging:

- Which recent requests hit a rule, fell back, missed completely, or errored?
- Which namespace, ruleset, and rule handled a request?
- Was the fallback caused by a ruleset miss or a rule miss?
- Did fallback forward the original request or return a configured response?
- What status and latency did a request produce?
- How can a user jump from a runtime symptom to the ruleset workspace that owns the behavior?
- How can a user replay a recent request shape through the simulation path to understand why it matched or missed?

Without a diagnostics surface, users can configure rules but still need logs or manual reproduction to understand live behavior.

## Solution

Redesign Runtime Board into a Runtime Diagnostics Dashboard.

The dashboard should combine aggregate runtime health with a recent request stream. It should show high-signal cards for traffic volume, match rate, fallback volume, errors, latency, and response status distribution. It should expose a recent runtime request list with method, path, namespace, status, duration, trace ID, matched ruleset, matched rule, fallback reason, and outcome.

Users should be able to filter and search recent runtime requests, jump directly from a matched request to the relevant ruleset workspace, and replay a stored request event through published simulation for diagnosis. The dashboard should stay aligned with the existing console/workbench UI direction and should not become another theme pass.

The backend should remain lightweight. Runtime diagnostics can be held in the existing in-memory observability component as a bounded recent-request buffer. This is sufficient for local MockServer operation and avoids schema or persistence work.

## User Stories

1. As a MockServer user, I want to see recent runtime requests, so that I can debug behavior without reading server logs.
2. As a MockServer user, I want each recent request to show method and path, so that I can recognize the request I am investigating.
3. As a MockServer user, I want each recent request to show namespace, so that I understand which fallback policy applies.
4. As a MockServer user, I want each recent request to show matched ruleset and matched rule, so that I can jump to the owning rule configuration.
5. As a MockServer user, I want fallback requests to show ruleset miss or rule miss, so that I know whether selector matching or rule condition matching failed.
6. As a MockServer user, I want runtime errors to be visually distinct from misses, so that I can separate configuration debugging from service failures.
7. As a MockServer user, I want response status and duration visible per request, so that I can triage slow or failing behavior quickly.
8. As a MockServer user, I want trace ID visible per request, so that I can correlate dashboard rows with logs and caller traces.
9. As a MockServer user, I want to search recent runtime requests by path, namespace, trace ID, ruleset ID, rule ID, and fallback reason, so that I can find the relevant request quickly.
10. As a MockServer user, I want to filter recent runtime requests by matched, fallback, unmatched, and error outcomes, so that I can focus on the kind of behavior I am debugging.
11. As a MockServer user, I want to filter recent runtime requests by namespace, so that I can isolate one mock domain.
12. As a MockServer user, I want fallback reason counts, so that I can tell whether most misses are ruleset miss or rule miss.
13. As a MockServer user, I want status code counts, so that I can tell whether runtime behavior is mostly successful, fallback, or failing.
14. As a MockServer user, I want top ruleset and top rule rankings to remain visible, so that I do not lose the existing aggregate insight.
15. As a MockServer user, I want to click a matched ruleset or rule and open the rules workspace, so that I can fix the behavior immediately.
16. As a MockServer user, I want the rules workspace to honor a rule deep link, so that a dashboard row can focus the relevant rule.
17. As a MockServer user, I want to replay a recent request through published simulation, so that I can inspect match explanations without manually rebuilding the event.
18. As a MockServer user, I want replay results to use the existing structured result inspector, so that simulation output stays consistent with the rules workspace.
19. As a MockServer user, I want the dashboard to remain readable on narrow viewports, so that runtime triage is usable outside a full desktop width.
20. As a MockServer user, I want empty runtime state to explain what is missing, so that a fresh server is understandable.
21. As a MockServer developer, I want runtime diagnostics derived through a tested view-model helper, so that UI filtering and classification do not live only in a Vue page.
22. As a MockServer developer, I want the in-memory recent request buffer to be bounded, so that diagnostics do not grow without limit.
23. As a MockServer developer, I want the existing metrics endpoint to stay backward-compatible, so that existing fields and consumers keep working.
24. As a MockServer maintainer, I want backend tests for runtime observations and recent diagnostics, so that future matching changes preserve the diagnostics contract.
25. As a MockServer maintainer, I want frontend tests for request outcome classification and filters, so that future dashboard changes preserve debugging behavior.

## Implementation Decisions

- Treat this as runtime diagnostics and information architecture work, not as a visual theme redesign.
- Keep the current Vue 3, TypeScript, Vite, Element Plus, Pinia, SCSS, and console/workbench direction.
- Extend the existing runtime metrics endpoint with additional backward-compatible fields rather than adding a separate diagnostics endpoint.
- Keep existing metrics fields for total requests, matched requests, unmatched requests, errors, average duration, ruleset matches, rule matches, and last matched rule timestamps.
- Add a bounded in-memory recent request buffer to the runtime observability component. This buffer should hold only a limited number of recent runtime observations.
- Each runtime observation should capture method, host, path, query, namespace, trace ID, status, duration, outcome, ruleset ID, rule ID, fallback reason, and an optional replayable event.
- Count fallback reasons and status codes as additional aggregate diagnostics.
- Do not persist runtime diagnostics to disk or database in this PRD.
- Do not count admin simulation calls as live runtime traffic; replay should use the existing published simulation path.
- Dashboard should use a pure TypeScript runtime diagnostics view-model helper for outcome labels, search text, filters, sorting, aggregate cards, and replay availability.
- Dashboard should preserve existing hit rankings but reposition them as supporting diagnostics rather than the primary surface.
- Dashboard request rows should use the ruleset workspace as the primary drill-down when a ruleset ID is available.
- Ruleset workspace should support a rule deep link query so runtime rows can focus the relevant rule.
- Replay should use the captured event shape when available and display the response through the existing structured result inspector.
- If a request cannot be replayed because no event is available, the UI should say so clearly instead of failing silently.
- The dashboard should use compact controls and stable dimensions so long paths and IDs remain inspectable without breaking layout.

## Testing Decisions

- Good tests should verify observable diagnostics behavior: recorded request shape, bounded buffer behavior, aggregate counts, filtering, classification, navigation eligibility, and replay eligibility.
- Backend tests should cover matched requests, fallback requests, error/unmatched requests, fallback reason counts, status code counts, and recent request retention.
- Frontend tests should cover runtime diagnostics view-model behavior: outcome classification, search, filters, namespace filtering, sorting, metrics, and replay availability.
- Existing server flow tests should continue to pass and should verify that legacy metrics fields remain present.
- Dashboard browser checks should cover desktop and narrow viewport readability, recent request rows, filters, ruleset drill-down action, replay entry, and empty state.
- Verification should include frontend tests, type-check, production build, backend Go tests, and browser checks.

## Out of Scope

- Persisting runtime diagnostics across process restarts.
- Building a full distributed tracing system.
- Adding authentication or permission changes.
- Recording unbounded request bodies or long-term traffic history.
- Replacing existing metrics storage with Prometheus or another external telemetry backend.
- Changing matching semantics, fallback runtime semantics, namespace defaults, or SDK decision semantics.
- Reworking the RuleManager or Workbench layout again beyond honoring a deep link.
- Building a full E2E test suite for every frontend route.

## Further Notes

This PRD follows the Ruleset and Namespace entry experience work. The goal is to close the loop from "configure rules" to "observe runtime behavior" to "jump back into the owning rule and fix it."

## Completion Notes

- Implemented bounded in-memory recent runtime diagnostics on the existing metrics endpoint.
- Rebuilt the dashboard around runtime health cards, searchable recent requests, status/fallback distributions, hit rankings, rule drill-down, and replay.
- Added rule deep-link handling in the rules workspace.
- Verified with unit tests, type-check, production build, Go tests, `git diff --check`, and Playwright desktop/narrow viewport checks.
