# Issue 02: Exact Traffic Stats And Bounded Indexed Filters

Status: completed
Blocked by: none

## Scope

- Replace sampled traffic stats with exact grouped stats for the active filter.
- Avoid unbounded indexed-field ID materialization.
- Preserve pagination behavior for event list results.
- Preserve current schema.

## Acceptance

- Traffic stats are exact for all rows matching filters.
- Indexed-field filters apply time/protocol/path/hash/value constraints before event lookup.
- Materialized indexed-filter event IDs are bounded.
- Tests cover the DAO behavior without requiring MySQL.

## Result

- Replaced sampled traffic stats with exact grouped stats.
- Added indexed-filter materialization cap.
- Preserved business-code filtering alongside resolved DB IDs for historical traffic queries.
- Added DAO unit coverage for exact stats propagation and DBID/code filter generation.
