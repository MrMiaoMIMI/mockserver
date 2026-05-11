# Runtime Recent Request Diagnostics

Status: done
Type: AFK

## Parent

.scratch/runtime-diagnostics-dashboard/PRD.md

## What to build

Add a bounded runtime diagnostics stream to the existing runtime metrics contract. Runtime requests should continue updating the existing aggregate metrics while also recording recent request summaries that explain method, path, namespace, status, duration, match state, fallback reason, trace ID, and linked ruleset/rule information.

## Acceptance criteria

- [x] Existing runtime metrics fields remain present and backward-compatible.
- [x] Runtime observations record recent request summaries for matched, fallback, unmatched, and error outcomes.
- [x] Recent request summaries include method, host, path, query, namespace, trace ID, status, duration, ruleset ID, rule ID, fallback reason, and an optional replayable event.
- [x] Runtime metrics include fallback reason counts and status code counts.
- [x] The recent request buffer is bounded and drops oldest entries first.
- [x] Backend tests cover matched observations, fallback observations, error/unmatched observations, aggregate counts, and bounded retention.
- [x] `go test ./...` passes.

## Blocked by

None - can start immediately.

## Verification

- `go test ./...`
- Runtime metrics smoke check against `localhost:18080` confirmed `recent_requests`, `fallback_reasons`, and `status_codes`.
