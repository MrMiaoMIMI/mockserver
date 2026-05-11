# Final docs and QA

Status: completed
Type: AFK

## Parent

.scratch/protocol-end-to-end-adaptation/PRD.md

## What to build

Complete documentation, issue status updates, and verification for the protocol end-to-end adaptation. The completed feature should clearly explain generic selectors, protocol-driven frontend authoring, SDK protocol adapters, and the HTTP runtime debug exception.

## Acceptance criteria

- [x] PRD and issue files reflect completed work.
- [x] User docs explain generic selectors for HTTP and cache.
- [x] SDK docs explain core versus protocol adapter responsibilities.
- [x] Repository search finds no stale `hosts/path_prefixes` selector examples in updated docs or frontend fixtures.
- [x] `go test ./...` passes.
- [x] Frontend build passes.

## Blocked by

- .scratch/protocol-end-to-end-adaptation/issues/04-protocol-driven-rule-authoring.md
- .scratch/protocol-end-to-end-adaptation/issues/05-sdk-protocol-adapters.md

