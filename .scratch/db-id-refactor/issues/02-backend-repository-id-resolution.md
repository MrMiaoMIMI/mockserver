# Issue 02: Backend Repository ID Resolution

Status: done

Blocked by: issues/01-schema-id-code-model.md

## Scope

Update DO/FMO models and repository logic to resolve external codes to numeric database identities.

## Tasks

- Rename model fields to `RuleSetCode`, `NamespaceCode`, `SnapshotCode`, `EventCode`, and `RuleCode` where they map to string database columns.
- Switch table `IdFieldName()` implementations to the auto-increment `id`.
- Query rulesets, namespaces, and snapshots by code through explicit indexed queries.
- Store current published snapshot references by numeric IDs.

## Acceptance

- Create, update, publish, rollback, list, and get flows work through external codes while the DB relationships use numeric IDs.
