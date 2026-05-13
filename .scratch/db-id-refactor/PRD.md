# PRD: Database ID Refactor

Status: done

## Background

Mockserver currently stores several user-facing identifiers as long indexed string fields, including ruleset, namespace, snapshot, traffic event, and rule identifiers. Some tables already have an auto-increment primary key, but repository code still treats business strings as the effective database identity. This increases index size, mixes API terminology with database relationships, and makes future high-volume traffic storage harder to optimize.

## Goals

- Use `BIGINT UNSIGNED` auto-increment `id` as the internal database identity.
- Store user-facing string identifiers as short `*_code` values in MySQL.
- Use numeric references for internal relationships and hot query indexes.
- Keep user-facing APIs ergonomic: users should still work with short readable identifiers, not large numeric database keys.
- Shorten generated snapshot identifiers so they do not duplicate ruleset code, version, and timestamp information already stored in dedicated columns.
- Add explicit length and character validation before data reaches MySQL.

## Non-Goals

- Preserve old database schema compatibility.
- Migrate legacy rows.
- Introduce foreign keys, stored procedures, triggers, or third-party dependencies.

## Requirements

1. Rule sets
   - Draft rulesets store `ruleset_code` as a short unique business code.
   - The table primary key `id` is the internal ruleset identity.
   - Code generation caps generated ruleset codes at 96 characters.
   - Manual ruleset codes are normalized and validated before persistence.

2. Namespaces
   - Namespaces store `namespace_code` as a short unique business code.
   - The table primary key `id` is the internal namespace identity.
   - Namespace codes are capped at 64 characters and remain readable for SDK/runtime configuration.

3. Snapshots
   - Published snapshots store `snapshot_code` as a short unique code.
   - Published snapshot rows reference rulesets by numeric `ruleset_id`.
   - The current-published table stores numeric `ruleset_id` and numeric `current_snapshot_id`.
   - Generated snapshot codes are short random codes, not `ruleset-version-time` concatenations.

4. Runtime traffic
   - Traffic rows store `event_code` as the external event identifier.
   - Traffic rows store numeric `namespace_id`, `ruleset_id`, and `snapshot_id` when those references are resolvable.
   - Traffic rows may also store short denormalized `namespace_code`, `ruleset_code`, `rule_code`, and `snapshot_code` for display and debugging.
   - Hot filters use numeric IDs where possible.
   - The traffic index table relates to traffic events only through numeric `traffic_event_id`.

5. MySQL design compliance
   - Tables use `ENGINE=InnoDB` and `utf8mb4_unicode_ci`.
   - Table names end with `_tab`; database name ends with `_db`.
   - Internal IDs are `BIGINT UNSIGNED`.
   - Fields are `NOT NULL` by default.
   - No foreign keys, triggers, enums, stored procedures, or nullable fields are introduced.

## Acceptance Criteria

- Schema DDL uses `*_code` for user-facing string identifiers and `*_id` for numeric references.
- Backend repositories no longer rely on long string IDs as database row identity.
- Snapshot codes are short and unique.
- Traffic event indexes no longer duplicate event string identifiers.
- Existing API/user workflows remain readable through short external identifiers.
- Backend tests pass.
- Frontend build passes.
