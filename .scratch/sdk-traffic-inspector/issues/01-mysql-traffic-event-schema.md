# 01. MySQL Traffic Event Schema

Status: completed

Blocked by: none

## Scope

Add guide-compliant MySQL tables and model types for persisted SDK decision traffic:

- `mockserver_traffic_event_tab`
- `mockserver_traffic_event_index_tab`

## Requirements

- Table names end with `_tab`.
- IDs use `BIGINT UNSIGNED`.
- Time fields use UNIX seconds.
- No nullable fields, foreign keys, enums, triggers, or stored routines.
- Store full protocol field values in `TEXT`, preview in `VARCHAR(255)`, and application-computed hash in `BIGINT UNSIGNED`.
- Do not add HTTP-only `status_code` to the main event table.

## Verification

- Schema unit test covers the new table names, field types, indexes, and forbidden patterns.

## Completion Notes

- Added `mockserver_traffic_event_tab` and `mockserver_traffic_event_index_tab` to runtime schema and manual schema docs.
- Added DO/FMO/BO models and repository plumbing.
- Added schema checks for table names, `TEXT` full values, `BIGINT UNSIGNED` hash, and absence of `status_code`.
