# Generic selector contract

Status: completed
Type: AFK

## Parent

.scratch/protocol-end-to-end-adaptation/PRD.md

## What to build

Replace HTTP-specific ruleset selectors with protocol field predicates. A completed slice should let HTTP and cache rulesets use the same selector shape, validate selector fields against the selected protocol, and explain selector matching with protocol field paths.

## Acceptance criteria

- [x] Ruleset selectors are represented as protocol field predicates instead of HTTP `hosts/path_prefixes`.
- [x] Empty selectors remain valid and match all rulesets with matching protocol and namespace.
- [x] Selector validation rejects fields not declared in `ProtocolSpec.Selectors`.
- [x] Selector validation rejects unsupported operators for selector fields.
- [x] HTTP selectors can match host equality and path prefix through the generic selector shape.
- [x] Cache selectors can match operation equality and key prefix through the generic selector shape.
- [x] Simulation explain output contains selector checks for generic selector predicates.
- [x] Focused Go tests cover HTTP and cache selector behavior.

## Blocked by

None - can start immediately.

