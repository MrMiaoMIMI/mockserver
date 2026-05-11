# Protocol-driven rule authoring

Status: completed
Type: AFK

## Parent

.scratch/protocol-end-to-end-adaptation/PRD.md

## What to build

Adapt the frontend rule authoring and simulation helpers to use protocol fields as the primary field source. HTTP and cache should both have useful condition fields, operators, and default simulation events.

## Acceptance criteria

- [x] Condition field options are derived from the selected protocol spec.
- [x] Operator options are derived from the selected field's effective operators.
- [x] HTTP preset shortcuts remain available as convenience, not as the only field source.
- [x] Cache condition authoring supports operation, key, ttl, and value paths.
- [x] Simulation default events are protocol-aware for HTTP and cache.
- [x] Frontend tests cover protocol-aware condition options and simulation defaults.

## Blocked by

- .scratch/protocol-end-to-end-adaptation/issues/03-protocol-driven-ruleset-settings.md

