# Forward decision for miss paths

Status: done
Type: AFK

## Parent

.scratch/mocksdk-package/PRD.md

## What to build

Extend the SDK decision path so ruleset miss and rule miss resolve to forward decisions when namespace fallback is configured as forward. The backend should resolve fallback but must not execute client-side forward behavior for the SDK decision endpoint. The SDK should parse the forward decision and make the fallback reason and forward metadata available to mockinject-style callers.

Covers user stories: 4, 13, 31, 32, 38, 41.

## Acceptance criteria

- [x] A ruleset miss returns a typed forward decision when the namespace ruleset miss fallback action is forward.
- [x] A rule miss returns a typed forward decision when the namespace rule miss fallback action is forward.
- [x] Forward decisions include fallback reason, trace metadata, and the forward timeout or effective forward policy returned by MockServer.
- [x] The SDK parses forward decisions distinctly from transport errors and mock response decisions.
- [x] The SDK decision endpoint does not perform server-side forwarding for forward decisions.
- [x] The existing direct runtime endpoint still performs server-side forward fallback as before.
- [x] Tests cover ruleset miss, rule miss, default namespace forward fallback, and preservation of existing runtime forwarding.
- [x] `go test ./...` passes.

## Blocked by

- .scratch/mocksdk-package/issues/01-published-hit-decision-tracer-bullet.md

## Comments

- Added forward decisions to the runtime SDK decision contract with `kind=forward`, `fallback=true`, `trace.fallback_reason`, and `forward.timeout_ms`.
- Reused published matching and extracted namespace fallback resolution so runtime can continue executing forward fallback while SDK decision returns a forward instruction without contacting upstream.
- Backend coverage verifies ruleset miss and rule miss forward decisions, default namespace forward policy, trace metadata, and that the SDK decision endpoint does not forward upstream requests.
- Existing runtime forwarding behavior remains covered by the new SDK/ruleset miss check and existing runtime rule miss/default namespace tests.
- SDK coverage verifies that the Go client parses forward decisions separately from response decisions.
- Verified with `go test ./...`.
