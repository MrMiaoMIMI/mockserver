# Runtime Diagnostics View Model

Status: done
Type: AFK

## Parent

.scratch/runtime-diagnostics-dashboard/PRD.md

## What to build

Create a testable frontend runtime diagnostics view model that turns runtime metrics and recent request records into dashboard cards, request rows, outcome labels, filters, sort order, status/fallback summaries, and replay or drill-down availability.

## Acceptance criteria

- [x] The view model classifies requests as matched, fallback, unmatched, or error.
- [x] The view model derives readable labels, tones, search text, status labels, duration labels, and navigation/replay affordances.
- [x] Users can search by path, namespace, trace ID, ruleset ID, rule ID, status, and fallback reason.
- [x] Users can filter by outcome and namespace.
- [x] Users can sort recent requests by newest, slowest, status, and outcome.
- [x] Unit tests cover classification, search, filtering, sorting, dashboard cards, fallback summaries, and replay availability.
- [x] Frontend tests and type-check pass.

## Blocked by

- .scratch/runtime-diagnostics-dashboard/issues/01-runtime-recent-request-diagnostics.md

## Verification

- `npm run test`
- `npm run type-check`
