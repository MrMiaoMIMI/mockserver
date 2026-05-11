# HTTP request fidelity and body replay

Status: done
Type: AFK

## Parent

.scratch/mocksdk-package/PRD.md

## What to build

Make the SDK useful for real injected HTTP flows by adding a request normalization path that converts an original HTTP request into the SDK decision request while preserving the original request for later forwarding. The normalized decision request should match MockServer runtime semantics for method, scheme, host, original host, path, query, headers, body, client IP, and trace ID.

Covers user stories: 2, 5, 6, 7, 25, 26, 27, 34, 42.

## Acceptance criteria

- [x] The SDK can build a decision request from an original HTTP request without requiring callers to hand-author the MockServer event shape.
- [x] The normalized request preserves method, scheme, host, original host, path, query, headers, raw body, parsed body where safe, client IP, and trace ID.
- [x] Reading the request body for normalization does not prevent injected code from forwarding the original request afterward.
- [x] JSON, plain text, empty body, and non-JSON body inputs are handled predictably.
- [x] Header copying is deterministic and handles hop-by-hop headers consistently for later helper usage.
- [x] Tests cover the normalized event shape and body replay behavior.
- [x] `go test ./...` passes.

## Blocked by

- .scratch/mocksdk-package/issues/01-published-hit-decision-tracer-bullet.md

## Comments

- Added SDK HTTP request normalization and `Client.DecideHTTP`.
- Normalization preserves request shape, trace metadata, parsed JSON body when safe, raw body, and restores request body replay through `Body` and `GetBody`.
- Added hop-by-hop header classification for helper usage.
- Verified JSON, plain text, empty body, header copying, query copying, client IP, host/original host, and replay behavior with SDK tests.
- Verified with `go test ./...`.
