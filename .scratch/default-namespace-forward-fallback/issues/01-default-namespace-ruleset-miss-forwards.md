# Default namespace ruleset miss forwards original request

Status: done
Type: AFK

## Parent

.scratch/default-namespace-forward-fallback/PRD.md

## What to build

When a runtime request uses the `default` namespace and no published ruleset selector matches it, MockServer should use forward fallback by default and proxy the original request to the upstream target. The behavior should be observable through the runtime response and through the admin namespace representation for `default`.

The completed slice should prove the default namespace no longer behaves like a synthetic 404 endpoint for ruleset miss. It should still preserve the existing fallback reason and fallback marker so users can distinguish pass-through fallback from a mock hit.

## Acceptance criteria

- [ ] A runtime request in the `default` namespace with no matching published ruleset forwards to an upstream server instead of returning the old default 404 response.
- [ ] The forwarded request preserves the original method, path, query string, request body, and useful request headers while still excluding hop-by-hop headers.
- [ ] The runtime response returns the upstream status, headers, and JSON body.
- [ ] The runtime response marks fallback and reports the fallback reason as `ruleset_miss`.
- [ ] Querying or listing namespaces shows the `default` namespace with forward fallback configured for ruleset miss and rule miss.
- [ ] Existing explicit response fallback behavior for custom namespaces remains available and covered by tests.

## Blocked by

None - can start immediately.

## Comments
