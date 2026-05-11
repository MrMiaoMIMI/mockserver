# Published hit decision tracer bullet

Status: done
Type: AFK

## Parent

.scratch/mocksdk-package/PRD.md

## What to build

Establish the first end-to-end Mock SDK decision path. A published HTTP ruleset hit should be evaluable through a runtime-safe SDK decision contract, and a public Go SDK client should be able to call that contract and return a typed mock response decision to mockinject-style callers.

This slice should prove the package and protocol shape without trying to cover every miss or fallback case yet. It should keep the existing direct runtime endpoint behavior unchanged.

Covers user stories: 1, 3, 12, 15, 16, 17, 19, 30, 34, 37, 39, 40, 41, 43.

## Acceptance criteria

- [x] A runtime-safe SDK decision endpoint can evaluate published rulesets and return a mock response decision for a matched HTTP request.
- [x] The decision contract includes response status, headers, body, matched state, ruleset ID, rule ID, and trace metadata.
- [x] A public Go SDK client can call the decision endpoint and expose the result as a typed mock response decision.
- [x] SDK public types do not import or expose MockServer internal business objects.
- [x] The existing direct runtime endpoint still returns matched mock responses as before.
- [x] Backend and SDK tests cover the published hit path with an in-process or in-memory HTTP server.
- [x] `go test ./...` passes.

## Blocked by

None - can start immediately.

## Comments

- Added the first SDK decision endpoint for published HTTP rule hits. It returns `data.decision.kind=response` through the standard success envelope.
- Added the public Go SDK package and client for requesting a decision from MockServer without importing internal packages.
- Backend coverage verifies a published hit decision and confirms the existing direct runtime endpoint still returns the same matched mock response.
- SDK coverage uses an in-memory HTTP server to verify endpoint path, request serialization, envelope decoding, response metadata, and trace metadata.
- Verified with `go test ./...`.
