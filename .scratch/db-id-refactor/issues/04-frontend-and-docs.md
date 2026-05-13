# Issue 04: Frontend And Docs Alignment

Status: done

Blocked by: issues/03-id-generation-validation-and-traffic.md

## Scope

Align frontend types, display helpers, docs, and tests with the shortened identifier model.

## Tasks

- Update frontend traffic index types after removing event string IDs from index rows.
- Keep user-visible copy focused on readable external identifiers.
- Update documentation and tests that assert the old snapshot code format.

## Acceptance

- Frontend build succeeds.
- User-facing screens still display ruleset, namespace, snapshot, event, trace, and rule identifiers clearly.
