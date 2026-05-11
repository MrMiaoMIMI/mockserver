# Dashboard Diagnostics Workflow

Status: done
Type: AFK

## Parent

.scratch/runtime-diagnostics-dashboard/PRD.md

## What to build

Redesign Runtime Board into a diagnostics workflow. Users should see aggregate runtime health, status/fallback distributions, a searchable recent request stream, drill-down links to the rules workspace, and a replay action that runs a captured request event through published simulation and shows structured results.

## Acceptance criteria

- [x] Dashboard shows high-signal cards for total requests, match rate, fallback requests, errors, and average duration.
- [x] Dashboard shows status code distribution, fallback reason distribution, ruleset hit ranking, and rule hit ranking.
- [x] Dashboard shows a recent request stream with method, path, namespace, outcome, status, duration, trace ID, ruleset ID, rule ID, and fallback reason.
- [x] Users can search, filter, and sort recent requests from the dashboard.
- [x] Users can open the rules workspace from a request with a ruleset ID.
- [x] Users can replay a request with a captured event and inspect the published simulation result.
- [x] Empty state and non-replayable rows are understandable.
- [x] Frontend tests, type-check, and build pass.

## Blocked by

- .scratch/runtime-diagnostics-dashboard/issues/01-runtime-recent-request-diagnostics.md
- .scratch/runtime-diagnostics-dashboard/issues/02-runtime-diagnostics-view-model.md

## Verification

- `npm run test`
- `npm run type-check`
- `npm run build`
- Playwright dashboard load, request stream, search, outcome filter, sort dropdown, replay, and console warning checks.
