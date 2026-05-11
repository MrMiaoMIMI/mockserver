# HTTP adapter helpers for applying decisions

Status: done
Type: AFK

## Parent

.scratch/mocksdk-package/PRD.md

## What to build

Add optional SDK helper APIs that make common HTTP mockinject adapters easy to write without hiding the core decision model. Callers should be able to apply a response decision to an HTTP response target. If helper forwarding is included in this slice, it must remain separate from decision retrieval so callers can choose how forwarding is performed.

Covers user stories: 18, 19, 23, 27.

## Acceptance criteria

- [x] The SDK provides a helper for applying a response decision to an HTTP response target without dropping status, headers, or body.
- [x] The helper layer remains optional; callers can still use the low-level decision client directly.
- [x] The core decision method does not automatically forward original requests.
- [x] If a forward helper is implemented, it uses preserved original request data and handles hop-by-hop headers predictably.
- [x] Helper tests cover response application for JSON, text, empty body, and multiple header values.
- [x] If a forward helper is implemented, tests cover forward request construction and skipped hop-by-hop headers.
- [x] `go test ./...` passes.

## Blocked by

- .scratch/mocksdk-package/issues/01-published-hit-decision-tracer-bullet.md
- .scratch/mocksdk-package/issues/02-forward-decision-for-miss-paths.md
- .scratch/mocksdk-package/issues/03-http-request-fidelity-and-body-replay.md
- .scratch/mocksdk-package/issues/05-response-fallback-decisions.md

## Comments

- Added optional `ApplyResponse` helper for writing SDK response decisions to an HTTP response target.
- Added optional `ForwardOriginal` helper for forwarding the preserved original request while skipping hop-by-hop headers.
- Core `Decide` and `DecideHTTP` still only retrieve decisions and do not automatically forward.
- Verified JSON, text, empty body, repeated headers, replayed request body, and skipped hop-by-hop headers.
- Verified with `go test ./...`.
