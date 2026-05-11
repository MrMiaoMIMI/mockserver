# Response fallback decisions

Status: done
Type: AFK

## Parent

.scratch/mocksdk-package/PRD.md

## What to build

Handle namespaces that intentionally return fallback responses for ruleset miss or rule miss. When namespace fallback is configured as response, the SDK decision endpoint should return a response decision rather than a forward decision, and the SDK should expose it through the same typed decision API as matched mock responses while preserving fallback metadata.

Covers user stories: 3, 13, 33, 35, 41.

## Acceptance criteria

- [x] A ruleset miss returns a typed response decision when the namespace ruleset miss fallback action is response.
- [x] A rule miss returns a typed response decision when the namespace rule miss fallback action is response.
- [x] Response fallback decisions include response status, headers, body, fallback state, fallback reason, and trace metadata.
- [x] The SDK exposes response fallback decisions through the same response decision path used for matched mock responses.
- [x] Tests cover ruleset miss response fallback and rule miss response fallback.
- [x] Existing namespace fallback response behavior for the direct runtime endpoint remains unchanged.
- [x] `go test ./...` passes.

## Blocked by

- .scratch/mocksdk-package/issues/01-published-hit-decision-tracer-bullet.md

## Comments

- Extended SDK decision generation so namespace response fallback returns `kind=response` with `fallback=true`.
- Response fallback decisions include status, headers, body, fallback reason, and trace metadata.
- Existing direct runtime response fallback behavior remains covered and unchanged.
- Verified ruleset miss response fallback and rule miss response fallback through controller tests.
- Verified with `go test ./...`.
