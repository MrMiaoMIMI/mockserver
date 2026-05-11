# HTTP runtime forwarding is an explicit debug exception

Status: done
Type: AFK

## Parent

.scratch/decision-only-forward-boundary/PRD.md

## What to build

Keep the backend decision endpoint side-effect free and make the remaining server-side forwarding path explicitly scoped to direct HTTP runtime debugging. A completed slice should demonstrate that SDK decision calls only return decisions, while direct HTTP runtime URLs may still forward original requests for local rule debugging.

Covers user stories: 6, 7, 8, 9, 10, 12, 13.

## Acceptance criteria

- [x] The SDK decision endpoint still returns `kind=forward` without opening upstream connections.
- [x] Direct HTTP runtime requests can still execute namespace forward fallback for debugging.
- [x] Backend code clearly scopes concrete server-side forwarding to HTTP runtime behavior.
- [x] Non-HTTP future integrations are not encouraged to reuse the HTTP runtime forwarding implementation as a generic fallback executor.
- [x] Focused controller or service tests cover both the decision-only path and the HTTP runtime exception.
- [x] `go test ./internal/...` passes.

## Blocked by

- .scratch/decision-only-forward-boundary/issues/01-mocksdk-forward-decision-without-forwarding-helper.md

## Comments

- Renamed fallback execution helpers to the HTTP runtime boundary and added a protocol guard for server-side forward execution.
- Added a service test proving non-HTTP runtime forward execution fails while `DecidePublished` still returns a forward decision for mockinject.
- Existing controller tests continue to prove the SDK decision endpoint does not forward and direct HTTP runtime fallback can forward for debugging.
- Verified with `go test ./internal/...`.
