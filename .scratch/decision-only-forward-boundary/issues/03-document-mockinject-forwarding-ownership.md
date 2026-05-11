# Document mockinject forwarding ownership

Status: done
Type: AFK

## Parent

.scratch/decision-only-forward-boundary/PRD.md

## What to build

Update user-facing documentation and examples so the tool-chain boundary is clear: MockServer and Mock SDK return decisions, while mockinject owns protocol-specific forwarding. Remove examples that call a Mock SDK forwarding helper.

Covers user stories: 9, 10, 14, 15.

## Acceptance criteria

- [x] Mock SDK docs describe `forward` as a decision that mockinject must execute.
- [x] Docs explicitly say direct HTTP runtime forwarding is only a runtime debugging convenience.
- [x] Minimal SDK examples no longer call a concrete forwarding helper.
- [x] Existing tool-chain docs do not imply that Mock SDK or generic MockServer own protocol-specific upstream forwarding.
- [x] A repository search no longer finds stale `ForwardOriginal` references.
- [x] `go test ./...` passes after docs and code changes.

## Blocked by

- .scratch/decision-only-forward-boundary/issues/01-mocksdk-forward-decision-without-forwarding-helper.md
- .scratch/decision-only-forward-boundary/issues/02-http-runtime-forwarding-is-explicit-debug-exception.md

## Comments

- Updated Mock SDK quickstart and tool-chain notes to make mockinject responsible for protocol-specific forwarding.
- Removed the `ForwardOriginal` example from docs.
- Verified stale helper references with repository search and final `go test ./...`.
