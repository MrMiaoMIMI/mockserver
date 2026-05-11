# Default namespace rule miss forwards original request

Status: done
Type: AFK

## Parent

.scratch/default-namespace-forward-fallback/PRD.md

## What to build

When a runtime request uses the `default` namespace, a published ruleset selector matches, and no rule inside that ruleset matches, MockServer should use forward fallback by default and proxy the original request to the upstream target.

The completed slice should prove that partial mock coverage stays pass-through for unmatched rule conditions while preserving the distinction between ruleset miss and rule miss.

## Acceptance criteria

- [ ] A published ruleset in the `default` namespace can be selected by selector while the request misses every rule inside it.
- [ ] That rule miss forwards to an upstream server instead of returning the old default 404 response.
- [ ] The forwarded request preserves the original method, path, query string, request body, and useful request headers while still excluding hop-by-hop headers.
- [ ] The runtime response returns the upstream status, headers, and JSON body.
- [ ] The runtime response marks fallback and reports the fallback reason as `rule_miss`.
- [ ] Existing mock-hit behavior for a matching rule in the same namespace still returns the configured mock response rather than forwarding.

## Blocked by

- .scratch/default-namespace-forward-fallback/issues/01-default-namespace-ruleset-miss-forwards.md

## Comments
