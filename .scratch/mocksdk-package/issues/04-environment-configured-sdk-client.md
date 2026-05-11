# Environment configured SDK client

Status: done
Type: AFK

## Parent

.scratch/mocksdk-package/PRD.md

## What to build

Complete SDK configuration resolution for mockinject-style callers. The SDK should support explicit options while allowing environment variables to take highest priority for mockserver host and namespace ID. It should validate required configuration, support timeouts and custom HTTP clients, and classify configuration and transport failures clearly.

Covers user stories: 8, 9, 10, 11, 20, 21, 22, 23, 24, 28, 29, 44.

## Acceptance criteria

- [x] SDK configuration supports mockserver host, namespace ID, timeout, and custom HTTP client options.
- [x] Environment variables override explicit SDK options for mockserver host and namespace ID.
- [x] The final resolved configuration is validated before decision calls are made.
- [x] Missing mockserver host, invalid host, invalid namespace, malformed endpoint responses, server errors, and transport failures produce distinguishable SDK errors.
- [x] Decision calls honor context cancellation and SDK timeout settings.
- [x] The SDK does not retry decision requests by default.
- [x] A test-friendly SDK interface or client abstraction is available for mockinject adapter tests.
- [x] Tests cover option merging, environment precedence, validation, timeout, cancellation, and transport error classification.
- [x] `go test ./...` passes.

## Blocked by

- .scratch/mocksdk-package/issues/01-published-hit-decision-tracer-bullet.md

## Comments

- Added SDK config resolution for mockserver host, namespace, timeout, and custom HTTP client.
- Added environment override support for `MOCKSERVER_HOST`, `MOCKSERVER_NAMESPACE_ID`, and `MOCKSERVER_TIMEOUT_MS`.
- Added SDK error kinds for config, transport, server, and decode failures.
- Added a small `Decider` interface for mockinject adapter tests.
- Verified option merging, environment precedence, validation, timeout, server status, malformed response, and custom HTTP client behavior.
- Verified with `go test ./...`.
