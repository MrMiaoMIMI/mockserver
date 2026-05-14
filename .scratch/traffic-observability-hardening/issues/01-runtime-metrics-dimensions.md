# Issue 01: Protocol-Aware Runtime Metrics

Status: completed
Blocked by: none

## Scope

- Add source/protocol/operation fields to runtime observations and recent records.
- Add fallback breakdowns by protocol, namespace, and operation.
- Populate suppressed SDK `ruleset_miss` metrics with protocol-aware event summary.

## Acceptance

- Suppressed SDK `ruleset_miss` is counted under source `sdk_decision`.
- Metrics can answer which protocol/namespace/operation produced `ruleset_miss`.
- Tests cover SPEX-like suppressed ruleset miss metrics.

## Result

- Added source/protocol/operation to runtime observation snapshots and recent records.
- Added fallback stats by protocol, namespace, and operation.
- Added controller coverage for SPEX-like SDK `ruleset_miss` metrics.
