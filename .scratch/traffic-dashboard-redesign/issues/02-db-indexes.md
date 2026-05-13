# Issue 02: Traffic Query Indexes

Status: done
Blocked by: 01-backend-summary-detail-contract.md

## Scope

- Update default schema and schema tests with the primary traffic timeline index.
- Follow the MySQL Database Design Guide naming and field rules.

## Acceptance

- `mockserver_traffic_event_tab` has `idx_traffic_source_event_time_id`.
- No foreign keys, stored routines, nullable fields, timestamp/datetime fields, or non-compliant index names are introduced.

## Result

- Added `idx_traffic_source_event_time_id (traffic_source, event_time, id)` to schema DDL and schema tests.
