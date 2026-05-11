# Protocol-driven ruleset settings

Status: completed
Type: AFK

## Parent

.scratch/protocol-end-to-end-adaptation/PRD.md

## What to build

Adapt the frontend ruleset settings flow to consume the protocol catalog. Users should choose a protocol and then configure selector predicates from the selected protocol's selector fields instead of fixed HTTP selector rows.

## Acceptance criteria

- [x] Frontend store loads and caches the protocol catalog.
- [x] Ruleset settings protocol options come from the protocol catalog.
- [x] Selector rows render field and operator options from the selected protocol.
- [x] Dynamic selector field paths can be extended when the protocol selector allows dynamic paths.
- [x] Rule list and rule detail pages summarize generic selector predicates.
- [x] Frontend tests cover selector conversion and summary output for HTTP and cache rulesets.

## Blocked by

- .scratch/protocol-end-to-end-adaptation/issues/02-effective-protocol-catalog.md

