# SDK quickstart docs and final integration QA

Status: done
Type: AFK

## Parent

.scratch/mocksdk-package/PRD.md

## What to build

Document and verify the completed Mock SDK package as a usable part of the mock tool chain. Users should be able to understand how MockServer, Mock SDK, and mockinject fit together, how environment overrides work, what each decision kind means, and how to run a minimal SDK integration path.

Covers user stories: 36, 43, 44, 45.

## Acceptance criteria

- [x] User-facing documentation explains the MockServer, Mock SDK, and mockinject tool-chain roles.
- [x] Documentation lists SDK environment variables, their precedence over explicit options, and expected values.
- [x] Documentation explains response decisions, forward decisions, matched metadata, fallback metadata, and common error classes.
- [x] A minimal SDK usage example demonstrates requesting a decision and applying a mock response or forwarding the original request.
- [x] Final integration QA covers at least one published hit, one forward decision, and one response fallback decision through the SDK path.
- [x] Documentation does not imply that mockinject itself is implemented in this repo.
- [x] Non-HTTP protocol support is explicitly documented as out of scope for the first SDK package.
- [x] `go test ./...` passes.
- [x] `git diff --check` passes.

## Blocked by

- .scratch/mocksdk-package/issues/01-published-hit-decision-tracer-bullet.md
- .scratch/mocksdk-package/issues/02-forward-decision-for-miss-paths.md
- .scratch/mocksdk-package/issues/03-http-request-fidelity-and-body-replay.md
- .scratch/mocksdk-package/issues/04-environment-configured-sdk-client.md
- .scratch/mocksdk-package/issues/05-response-fallback-decisions.md
- .scratch/mocksdk-package/issues/06-http-adapter-helpers-for-applying-decisions.md

## Comments

- Added `docs/mocksdk.md` and linked it from the README.
- Documentation covers tool-chain roles, environment precedence, decision endpoint, response/forward decision semantics, error classes, a minimal example, and HTTP-only scope.
- Added an end-to-end SDK client test against an in-process MockServer handler that covers published hit, forward decision, and response fallback decision.
- Verified with `go test ./...` and `git diff --check`.
