# Issue 01: Schema ID/Code Model

Status: done

Blocked by: none

## Scope

Refactor MySQL DDL so string business identifiers use `*_code` and numeric relationships use `BIGINT UNSIGNED *_id`.

## Tasks

- Update `docs/db_schema.sql`.
- Update runtime bootstrap schema in `internal/dao/schema.go`.
- Remove string event identifier from traffic index table.
- Ensure index names follow `idx_<fields>`.

## Acceptance

- Schema has no user-facing string business fields named `ruleset_id`, `namespace_id`, `snapshot_id`, `event_id`, or `rule_id`, except external `trace_id`.
- Internal references use unsigned integer ID fields.
